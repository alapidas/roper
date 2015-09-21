/*
The filesystem package provides a unified interface to a real backed filesystem.
An example use case for this might be a server that wants to provide CRUD access
to a filesystem.

Here's an example of how you might use this package.
The first step is to create a Filesystemer object based on the absolute
filesystem path you want to represent (using a provided constructor):
	fs := NewPassThroughFilesystem('/my/abs/path', false)
At this point, all interaction with the Filesystemer should be relative to the
root path.  For example, to check if /my/abs/path/lower/dir exists, you'd do
something like:
	fs.PathExists('lower/dir') // false
If it did not exist, and you wanted to make a director there, you could do:
	fs.MkPath('lower/dir')
If you wanted to then make a file in this path, you might do:
	fileBytes, err := ioutil.ReadFile('/tmp/roflcopter.jpeg')
	fs.MkFile('lower/dir/roflcopter.jpeg', fileBytes)
Then to retrieve the file, you'd do:
	myFiler, err := fs.GetFile('lower/dir/roflcopter.jpeg')
To remove the newly created file, this would work:
	fs.RmPath('lower/dir/roflcopter.jpeg')
	// Alternatively, recursively delete the whole directory
	// fs.RmPath('lower')
*/
package filesystem

import (
	"errors"
	"fmt"
	trie "github.com/tchap/go-patricia/patricia"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var (
	ErrFileNoExist     = errors.New("file does not exist")
	ErrPathNoExist     = errors.New("path does not exist")
	ErrFileExists      = errors.New("file already exists")
	ErrPathNotAbsolute = errors.New("path is not absolute")
)

// FSFiler knows everything that a normal os.FileInfo does, but it also
// has access to the underlying file.
type FSFiler interface {
	os.FileInfo
	File() []byte
}

// A FSFile object contains all the information that you would normally find
// in a os.FileInfo object, plus the file itself in a byte slice.
type FSFile struct {
	os.FileInfo
	file []byte
}

var _ FSFiler = (*FSFile)(nil)

// File returns the file stored in a FSFile
func (fsfile *FSFile) File() []byte {
	return fsfile.file
}

// Filesystemer is the base interface that any filesystem should implement.
type Filesystemer interface {
	//ArbitraryAction()
	PathExists(path string) bool
	MkPath(path string) error
	RmPath(path string) error // Applies to both paths + files
	MkFile(path string, content []byte) error
	GetFile(path string) (FSFiler, error)
	RootPath() string
}

// A TransientFilesystemer represents an in-memory representation of a filesystem.
type TransientFilesystemer interface {
	Filesystemer
	Sync() error
}

// A transient filesystem has an in-memory representation of an underlying
// filesystem, for fast access.  The in-memory structure will be updated
// when paths are added via the object.  When the filesystem is changed
// outside of this API, Sync() should be called to bring the in-memory
// representation and actual disk contents back into sync.
type TransientFilesystem struct {
	trie     *trie.Trie // not thread safe, and not embedded
	rootPath string
	lock     sync.RWMutex
}

var _ TransientFilesystemer = (*TransientFilesystem)(nil)

// Sync brings the in-memory struct and underlying file system into sync.
// Walk the entire tree and recreate it from scratch.
func (fs *TransientFilesystem) Sync() error {
	newTrie := trie.NewTrie()
	err := filepath.Walk(fs.RootPath(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fi, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !newTrie.Insert(trie.Prefix(path), fi) {
			return fmt.Errorf("Path %s already exists in tree", path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.trie = newTrie
	return nil
}

// NewTransientFilesystem returns a new TransientFilesystem.  Expects an
// absolute path to the location on disk.  Returns an error if the path cannot
// be created, or if a file already exists in that location.
// If create is true, then the directory will be created if it does not yet exist.
func NewTransientFilesystem(absPath string, create bool) (*TransientFilesystem, error) {
	if !filepath.IsAbs(absPath) {
		return nil, ErrPathNotAbsolute
	}
	fi, err := os.Stat(absPath)
	// short circuit for create
	if create && os.IsNotExist(err) {
		if err = os.MkdirAll(absPath, 0755); err != nil {
			return nil, err
		}
		fs := &TransientFilesystem{trie: trie.NewTrie(), rootPath: absPath}
		if err = fs.Sync(); err != nil {
			return nil, err
		}
		return fs, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Unable to stat path %s", absPath)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("Specified path %s is not a directory", absPath)
	}
	fs := &TransientFilesystem{trie: trie.NewTrie(), rootPath: absPath}
	if err = fs.Sync(); err != nil {
		return nil, err
	}
	return fs, nil
}

// RootPath returns the real absolute path of the filesystem
func (fs *TransientFilesystem) RootPath() string {
	return fs.rootPath
}

// PathExists checks to see that the given path exists.
// The path argument is expected to be relative to the rootPath of the
// TransientFilesystem.  Works for files and directories.
func (fs *TransientFilesystem) PathExists(path string) bool {
	absPath := realPath(fs, path)
	fs.lock.RLock()
	defer fs.lock.RUnlock()
	return fs.trie.Match(trie.Prefix(absPath))
}

// MkPath makes a new directory under the current rootPath.  The path variable
// is expected to be relative to the rootPath.  The diectory will both be added
// to the internal representation AND be added to the real filesystem.  Fails
// if the  specified path is a file, returns nil if the specified path already
// exists.
func (fs *TransientFilesystem) MkPath(path string) error {
	absPath := realPath(fs, path)

	// actually make path
	if absPath == fs.RootPath() {
		return nil
	}
	fi, err := os.Stat(absPath)
	if err != nil && !os.IsNotExist(err) {
		return ErrFileNoExist
	}
	if fi != nil && !fi.IsDir() {
		return ErrFileExists
	}
	doTrieGet := func() (bool, error) {
		fs.lock.RLock()
		defer fs.lock.RUnlock()
		if fs.trie.Match(trie.Prefix(absPath)) {
			if fs.trie.Get(trie.Prefix(absPath)).(os.FileInfo).IsDir() {
				return true, nil
			} else {
				return false, fmt.Errorf("Specified path %s exists in memory as a file, but does not exist on disk", path)
			}
		}
		return false, nil
	}
	existsInMem, err := doTrieGet()
	if err != nil {
		return err
	} else if existsInMem {
		return nil
	}
	if err = os.MkdirAll(absPath, 0755); err != nil {
		return err
	}
	fi, err = os.Stat(absPath)
	if err != nil {
		return err
	}

	// insert into trie
	insertIntoTrie := func() error {
		fs.lock.Lock()
		defer fs.lock.Unlock()
		if success := fs.trie.Insert(trie.Prefix(absPath), fi); !success {
			return fmt.Errorf("path %s already exists in memory", path)
		}
		return nil
	}
	if err := insertIntoTrie(); err != nil {
		// recover
		if rmErr := os.RemoveAll(absPath); rmErr != nil {
			// wow this is bad, failure inserting into the trie, AND failure
			// cleaning up after ourselves
			return fmt.Errorf("failure creating path %s in memory: %s, and failure cleaning up alrready created path %s: %s", path, err.Error(), path, rmErr.Error())
		}
		return err
	}
	return nil
}

// RmPath recursively removes a path from the filesystem.  The path argument
// can be either a file or directory relative to the rootPath.  Returns an
// error if the path does not exist, or if there was an error removing
// part of the path.
func (fs *TransientFilesystem) RmPath(path string) error {

	absPath := realPath(fs, path)
	switch path {
	case "/", "":
		return fmt.Errorf("unwilling to delete root filesystem path %s", fs.RootPath())
	default:
	}

	doGet := func() interface{} {
		fs.lock.RLock()
		defer fs.lock.RUnlock()
		return fs.trie.Get(trie.Prefix(absPath))
	}
	item := doGet()
	if item == nil {
		return ErrPathNoExist
	}
	if !item.(os.FileInfo).IsDir() {
		return ErrFileExists
	}
	// Make this a function so we can release the mutex ASAP
	doTrieDelete := func() error {
		fs.lock.Lock()
		defer fs.lock.Unlock()
		if deleted := fs.trie.Delete(trie.Prefix(absPath)); !deleted {
			return fmt.Errorf("path %s does not exist in memory", path)
		}
		return nil
	}
	if err := doTrieDelete(); err != nil {
		return err
	}
	if err := os.RemoveAll(absPath); err != nil {
		// put item back in the tree
		// eat the error here
		fs.trie.Insert(trie.Prefix(absPath), item)
		// possibly in an unknown state at this point.  whomp whomp.
		return err
	}
	return nil
}

// MkFile makes a file consisting of the provided content at the given path.
// If the path does not already exist, an error will be returned.  If the file
// already exists, an error will be returned.
func (fs *TransientFilesystem) MkFile(path string, content []byte) error {
	if !fs.PathExists(filepath.Dir(path)) {
		return ErrPathNoExist
	} else if fs.PathExists(path) {
		return ErrFileExists
	}

	absPath := realPath(fs, path)
	if err := ioutil.WriteFile(absPath, content, 0644); err != nil {
		return err
	}
	doInsertTrie := func() error {
		fs.lock.Lock()
		defer fs.lock.Unlock()
		fi, err := os.Stat(absPath)
		if err != nil {
			return err
		}
		if success := fs.trie.Insert(trie.Prefix(absPath), fi); !success {
			// path already exists in trie
			if err := os.RemoveAll(absPath); err != nil {
				// this is bad
				return fmt.Errorf("file %s already exists in memory and error removing file %s: %s", path, path, err)
			}
			// technically this is a little misleading, as it refers to the
			// trie, not what's on disk, but c'est la vie
			return ErrFileExists
		}
		return nil
	}
	if err := doInsertTrie(); err != nil {
		return err
	}
	return nil
}

// GetFile gets the file at path and returns a FSFiler object, which both
// describes the file and has a pointer to a copy of the file in a byte slice.
// If the files does not exist or path is actually a directory, an error is returned.
func (fs *TransientFilesystem) GetFile(path string) (FSFiler, error) {
	absPath := realPath(fs, path)
	if !fs.PathExists(path) {
		return nil, ErrFileNoExist
	} else if fi, err := os.Stat(realPath(fs, path)); err != nil || fi.IsDir() {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Specified path %s is a directory", realPath(fs, path))
	}
	item := fs.trie.Get(trie.Prefix(absPath))
	if item == nil {
		return nil, ErrPathNoExist
	}
	fileBytes, _ := ioutil.ReadFile(absPath)
	fi, _ := os.Stat(realPath(fs, path))
	return &FSFile{FileInfo: fi, file: fileBytes}, nil
}

// A passthrough file system doesn't have an internal representation of the
// actual filesystem - it is merely a proxy to the actual filesystem.  The
// penalty of disk access is paid, but it is never out of sync with the actual
// contents of the filesystem.  This implementation does not currently deal
// with file permissions.  That is, to say, files/direcories created via this
// implementation will be subject to ownership/permissions rules enforced by the kernel.
// If you want to be able to set custom ownership or permissions, provide your own
// implementation.
type PassThroughFilesystem struct {
	rootPath string
}

var _ Filesystemer = (*PassThroughFilesystem)(nil)

// Create a new PassThroughFileSystem, given an absolute path to a directory.
// Will fail if the given path is inaccessible for any reason (including its
// inexistence), or if the path is not a directory.  If create is true, then
// the directory will be created if it does not yet exist.
func NewPassThroughFilesystem(absPath string, create bool) (*PassThroughFilesystem, error) {
	if !filepath.IsAbs(absPath) {
		return nil, ErrPathNotAbsolute
	}
	fi, err := os.Stat(absPath)
	// short circuit for create
	if create && os.IsNotExist(err) {
		if err = os.MkdirAll(absPath, 0755); err != nil {
			return nil, err
		}
		return &PassThroughFilesystem{rootPath: absPath}, nil
	}
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, ErrFileExists
	}
	return &PassThroughFilesystem{rootPath: absPath}, nil
}

// RootPath returns the real absolute path of the filesystem
func (fs *PassThroughFilesystem) RootPath() string {
	return fs.rootPath
}

// realPath returns the absolute path of the PassThroughFilesystem's
// rootPath plus the given relPath.
func realPath(fs Filesystemer, relPath string) string {
	tgtPath := filepath.Join(fs.RootPath(), relPath)
	return filepath.Clean(tgtPath)
}

// PathExists checks to see that the given path exists.
// The path argument is expected to be relative to the rootPath of the
// PassThroughFilesystem.  Works for files and directories.
func (fs *PassThroughFilesystem) PathExists(path string) bool {
	_, err := os.Stat(realPath(fs, path))
	return !os.IsNotExist(err)
}

// MkPath makes a new directory under the current rootPath.  The path variable
// is expected to be relative to the rootPath.  Returns nil if the path already
// exists as a dir, and an error if the path already exists and is a file.
func (fs *PassThroughFilesystem) MkPath(path string) error {
	absPath := realPath(fs, path)
	if absPath == fs.rootPath {
		return nil
	}
	fi, err := os.Stat(absPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if fi != nil && !fi.IsDir() {
		return ErrFileExists
	}
	return os.MkdirAll(absPath, 0755)
}

// rmPathOnDisk removes the specified path from disk.  The given path is
// relative to the root of the associated Filesystemer object.
func rmPathOnDisk(fs Filesystemer, path string) error {
	absPath := realPath(fs, path)
	switch path {
	case "/", "":
		return fmt.Errorf("unwilling to delete root filesystem path %s", fs.RootPath())
	default:
	}
	if !fs.PathExists(path) {
		return ErrPathNoExist
	}
	return os.RemoveAll(absPath)
}

// RmPath recursively removes a path from the filesystem.  The path argument
// can be either a file or directory relative to the rootPath.  Returns an
// error if the path does not exist, or if there was an error removing
// part of the path.
func (fs *PassThroughFilesystem) RmPath(path string) error {
	return rmPathOnDisk(fs, path)
}

// MkFile makes a file consisting of the provided content at the given path.
// If the path does not already exist, an error will be returned.  If the file
// already exists, an error will be returned.
func (fs *PassThroughFilesystem) MkFile(path string, content []byte) error {
	if !fs.PathExists(filepath.Dir(path)) {
		return ErrPathNoExist
	} else if fs.PathExists(path) {
		return ErrFileExists
	}

	return ioutil.WriteFile(realPath(fs, path), content, 0644)
}

// GetFile gets the file at path and returns a FSFile object, which both
// describes the file and has a pointer to a copy of the file in a byte slice.
// If the files does not exist or path is actually a directory, an error is returned.
func (fs *PassThroughFilesystem) GetFile(path string) (FSFiler, error) {
	if !fs.PathExists(path) {
		return nil, ErrFileNoExist
	} else if fi, err := os.Stat(realPath(fs, path)); err != nil || fi.IsDir() {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Specified path %s is a directory", realPath(fs, path))
	}
	fileBytes, _ := ioutil.ReadFile(realPath(fs, path))
	fi, _ := os.Stat(realPath(fs, path))
	return &FSFile{FileInfo: fi, file: fileBytes}, nil
}

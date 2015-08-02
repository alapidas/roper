/*
The filesystem package provides a unified interface to a real backed filesystem.
An example use case for this might be a server that wants to provide CRUD access
to a filesystem.
*/
package filesystem

import (
	"fmt"
	trie "github.com/tchap/go-patricia/patricia"
	"io/ioutil"
	"os"
	"path/filepath"
)

// A FSFiler knows everything that a normal os.FileInfo does, but it also
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

func (fsfile *FSFile) File() []byte {
	return fsfile.file
}

// A TransientFilesystemer represents an in-memory representation of a filesystem.
type TransientFilesystemer interface {
	Filesystemer
	Sync() error
}

// Filesystemer is the base interface that any filesystem should implement.
type Filesystemer interface {
	PathExists(path string) bool
	MkPath(path string) error
	RmPath(path string) error // Applies to both paths + files
	MkFile(path string, content []byte) error
	GetFile(path string) (FSFiler, error)
}

type TransientFilesystem struct {
	*trie.Trie
	RootPath string
}

//var _ TransientFilesystemer = (*TransientFilesystem)(nil)

// A passthrough file system doesn't have an internal representation of the
// actual filesystem - it is merely a proxy to the actual filesystem.  The
// penalty of disk access is paid, but it is never out of sync with the actual
// contents of the filesystem.  This implementation does not currently deal
// with file permissions.  That is, to say, files/direcories created via this
// implementation will be subject to ownership/permissions rules enforced by the kernel.
// If you want to be able to set custom ownership or permissions, provide your own
// implementation.
type PassThroughFilesystem struct {
	RootPath string
}

var _ Filesystemer = (*PassThroughFilesystem)(nil)

// Create a new PassThroughFileSystem, given an absolute path to a directory.
// Will fail if the given path is inaccessible for any reason (including its
// inexistence), or is the path is not a directory.
func NewPassThroughFilesystem(path string) (*PassThroughFilesystem, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to stat path %s", path)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("Specified path %s is not a directory", path)
	}
	return &PassThroughFilesystem{RootPath: path}, nil
}

// realPath returns the absolute path of the PassThroughFilesystem's
// RootPath plus the given relPath.
func (fs *PassThroughFilesystem) realPath(relPath string) string {
	tgtPath := filepath.Join(fs.RootPath, relPath)
	return filepath.Clean(tgtPath)
}

// PathExists checks to see that the given path exist.
// The path argument is expected to be relative to the RootPath of the
// PassThroughFilesystem.  Works for files and directories.
func (fs *PassThroughFilesystem) PathExists(path string) bool {
	_, err := os.Stat(fs.realPath(path))
	return !os.IsNotExist(err)
}

// MkPath makes a new directory under the current RootPath.  The path variable
// is expected to be relative to the RootPath
func (fs *PassThroughFilesystem) MkPath(path string) error {
	absPath := fs.realPath(path)
	if absPath == fs.RootPath {
		return nil
	}
	return os.MkdirAll(absPath, 0755)
}

// RmPath recursively removes a path from the filesystem.  The path argument
// can be either a file or directory relative to the RootPath.  Returns an
// error if the path does not exist, or if there was an error removing
// part of the path.
func (fs *PassThroughFilesystem) RmPath(path string) error {
	absPath := fs.realPath(path)
	switch path {
	case "/", "":
		return fmt.Errorf("Unwilling to delete root filesystem path %s", fs.RootPath)
	default:
	}
	if !fs.PathExists(path) {
		return fmt.Errorf("Specified path %s does not exist", absPath)
	}
	return os.RemoveAll(absPath)
}

// MkFile makes a file consisting of the provided content at the given path.
// If the path does not already exist, an error will be returned.  If the file
// already exists, an error will be returned.
func (fs *PassThroughFilesystem) MkFile(path string, content []byte) error {
	if !fs.PathExists(filepath.Dir(path)) {
		return fmt.Errorf("Parent path %s does not exist", fs.realPath(filepath.Dir(path)))
	} else if fs.PathExists(path) {
		return fmt.Errorf("File %s already exists", fs.realPath(path))
	}

	return ioutil.WriteFile(fs.realPath(path), content, 0666)
}

// GetFile gets the file at path and returns a FSFile object, which both
// describes the file and has a pointer to a copy of the file in a byte slice.
// If the files does not exist or path is actually a directory, an error is returned.
func (fs *PassThroughFilesystem) GetFile(path string) (FSFiler, error) {
	if !fs.PathExists(path) {
		return nil, fmt.Errorf("File %s does not exist", fs.realPath(path))
	} else if fi, err := os.Stat(fs.realPath(path)); err != nil || fi.IsDir() {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Specified path %s is a directory", fs.realPath(path))
	}
	fileBytes, _ := ioutil.ReadFile(fs.realPath(path))
	fi, _ := os.Stat(fs.realPath(path))
	return &FSFile{FileInfo: fi, file: fileBytes}, nil
}

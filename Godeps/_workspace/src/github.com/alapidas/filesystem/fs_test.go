package filesystem

import (
	trie "github.com/tchap/go-patricia/patricia"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type TheSuite struct {
	tmpdir string
}

var _ = Suite(&TheSuite{})

// Create a temporary directory for every test to use
func (suite *TheSuite) SetUpTest(c *C) {
	suite.tmpdir = c.MkDir()
}

// helpers
func (suite *TheSuite) mkPTFS() *PassThroughFilesystem {
	fs, _ := NewPassThroughFilesystem(suite.tmpdir, false)
	return fs
}

func (suite *TheSuite) mkTFS() *TransientFilesystem {
	fs, _ := NewTransientFilesystem(suite.tmpdir, false)
	return fs
}

func (suite *TheSuite) mkDirs(c *C, dir string) map[string]string {
	fileMap := make(map[string]string)
	f1, err := ioutil.TempFile(suite.tmpdir, "f1")
	if err != nil {
		c.Error(err)
	}
	fileMap["f1"] = f1.Name()
	d1 := filepath.Join(suite.tmpdir, "d1")
	fileMap["d1"] = d1
	os.MkdirAll(d1, 0755)
	f2, err := ioutil.TempFile(d1, "f2")
	if err != nil {
		c.Error(err)
	}
	fileMap["f2"] = f2.Name()
	f3, err := ioutil.TempFile(d1, "f3")
	if err != nil {
		c.Error(err)
	}
	fileMap["f3"] = f3.Name()
	return fileMap
}

// Log the whole trie
func (suite *TheSuite) TrieLog(c *C, fs *TransientFilesystem) {
	fs.trie.VisitSubtree(trie.Prefix(""), func(prefix trie.Prefix, item trie.Item) error {
		c.Logf("%q: %v\n", prefix, item)
		return nil
	})
}

// tests
func (suite *TheSuite) TestPassThroughFilesystemerBadPath(c *C) {
	_, err := NewPassThroughFilesystem("/thispathbetternotexist", false)
	c.Assert(err, FitsTypeOf, &os.PathError{})
}

func (suite *TheSuite) TestPassThroughFilesystemerFileNotDir(c *C) {
	f, err := ioutil.TempFile(suite.tmpdir, "blah")
	if err != nil {
		c.Fatalf("Failed to make temp file in dir %s", suite.tmpdir)
	}
	_, err = NewPassThroughFilesystem(f.Name(), false)
	c.Assert(err, Equals, ErrFileExists)
}

func (suite *TheSuite) TestPassThroughFilesystemerCreate(c *C) {
	fs := suite.mkPTFS()
	c.Assert(fs.rootPath, Equals, suite.tmpdir)
	newPath := filepath.Join(suite.tmpdir, "under")
	fs, err := NewPassThroughFilesystem(newPath, false)
	c.Assert(err, NotNil)
	fs, err = NewPassThroughFilesystem(newPath, true)
	c.Assert(err, IsNil)
	c.Assert(fs.rootPath, Equals, newPath)
	fs, err = NewPassThroughFilesystem("asdasdb", true)
	c.Assert(err, Equals, ErrPathNotAbsolute)
}

func (suite *TheSuite) TestPassThroughFilesystemer_realPath(c *C) {
	fs := suite.mkPTFS()
	f, _ := ioutil.TempFile(fs.rootPath, "apples")
	f.Close()
	tmpfileName := filepath.Base(f.Name())
	absPath := realPath(fs, tmpfileName)
	c.Assert(absPath, Equals, filepath.Clean(filepath.Join(fs.rootPath, tmpfileName)))
}

func (suite *TheSuite) TestPassThroughFilesystemer_PathExists(c *C) {
	fs := suite.mkPTFS()
	f, _ := ioutil.TempFile(fs.rootPath, "apples")
	f.Close()
	tmpfileName := filepath.Base(f.Name())
	c.Assert(fs.PathExists(tmpfileName), Equals, true)
	c.Assert(fs.PathExists("applesssss"), Equals, false)
}

func (suite *TheSuite) TestPassThroughFilesystemer_MkPath(c *C) {
	fs := suite.mkPTFS()
	err := fs.MkPath("ab/cd")
	c.Assert(err, IsNil)
	c.Assert(fs.PathExists("ab/cd"), Equals, true)
	fs = suite.mkPTFS()
	err = fs.MkPath("/ab/cd")
	c.Assert(fs.PathExists("/ab/cd"), Equals, true)
	c.Assert(err, IsNil)
	err = fs.MkPath("")
	c.Assert(err, IsNil)
	err = fs.MkFile("here", []byte{})
	c.Assert(err, IsNil)
	err = fs.MkPath("here")
	c.Assert(err, Equals, ErrFileExists)
}

func (suite *TheSuite) TestPassThroughFilesystemer_RmPath(c *C) {
	fs := suite.mkPTFS()
	fs.MkPath("abc/123")
	c.Assert(fs.PathExists("abc/123"), Equals, true)
	err := fs.RmPath("abc/123")
	c.Assert(err, IsNil)
	err = fs.RmPath("abc/notexist")
	c.Assert(err, Equals, ErrPathNoExist)
	err = fs.RmPath("")
	c.Assert(err, NotNil)
	err = fs.RmPath("/")
	c.Assert(err, NotNil)

}

func (suite *TheSuite) TestPassThroughFilesystemer_MkFile(c *C) {
	fs := suite.mkPTFS()
	err := fs.MkFile("this/path/here", []byte{})
	c.Assert(err, Equals, ErrPathNoExist) //parent path does not exist
	err = fs.MkFile("here", []byte{})
	c.Assert(err, IsNil)

	cwd, _ := os.Getwd()
	fileBytes, _ := ioutil.ReadFile(filepath.Join(cwd, "thetestfile.txt"))
	fs.MkPath("myDir")
	err = fs.MkFile("myDir/theFile", fileBytes)
	c.Assert(err, IsNil)
	absPath := realPath(fs, "myDir/theFile")
	newBytes, _ := ioutil.ReadFile(absPath)
	c.Assert(fileBytes, DeepEquals, newBytes)

	err = fs.MkFile("myDir/theFile", fileBytes)
	c.Assert(err, Equals, ErrFileExists) //file already exists
}

func (suite *TheSuite) TestPassThroughFilesystemer_GetFile(c *C) {
	fs := suite.mkPTFS()
	_, err := fs.GetFile("dat/path") //file doesn't exist
	c.Assert(err, Equals, ErrFileNoExist)
	err = fs.MkPath("ab/cd")
	_, err = fs.GetFile("ab/cd") //actually a dir
	c.Assert(err, NotNil)

	cwd, _ := os.Getwd()
	fileBytes, _ := ioutil.ReadFile(filepath.Join(cwd, "thetestfile.txt"))
	fs.MkPath("myDir")
	fs.MkFile("myDir/theFile", fileBytes)
	c.Log("MADE IT HERE")

	newFile, err := fs.GetFile("myDir/theFile")
	c.Assert(err, IsNil)
	c.Assert(fileBytes, DeepEquals, newFile.File())

}

func (suite *TheSuite) TestTransientFilesystem_Create(c *C) {
	_, err := NewTransientFilesystem("asdasda", false)
	c.Assert(err, Equals, ErrPathNotAbsolute)
}

func (suite *TheSuite) TestTransientFilesystem_Sync(c *C) {
	fs := suite.mkTFS()
	fileMap := suite.mkDirs(c, suite.tmpdir)
	if err := fs.Sync(); err != nil {
		c.Error(err)
	}
	c.Assert(fs.trie.Get(trie.Prefix("aaaaaaaaaaa")), IsNil)
	c.Assert(fs.trie.Get(trie.Prefix(fs.rootPath)), NotNil)
	c.Assert(fs.trie.Get(trie.Prefix(fileMap["d1"])), NotNil)
	c.Assert(fs.trie.Get(trie.Prefix(fileMap["f1"])), NotNil)
	c.Assert(fs.trie.Get(trie.Prefix(fileMap["f2"])), NotNil)
	c.Assert(fs.trie.Get(trie.Prefix(fileMap["f3"])), NotNil)
}

func (suite *TheSuite) TestTransientFilesystemerCreate(c *C) {
	fileMap := suite.mkDirs(c, suite.tmpdir)
	fs := suite.mkTFS()
	c.Assert(fs.trie.Get(trie.Prefix("aaaaaaaaaaa")), IsNil)
	c.Assert(fs.trie.Get(trie.Prefix(fs.rootPath)), NotNil)
	c.Assert(fs.trie.Get(trie.Prefix(fileMap["d1"])), NotNil)
	c.Assert(fs.trie.Get(trie.Prefix(fileMap["f1"])), NotNil)
	c.Assert(fs.trie.Get(trie.Prefix(fileMap["f2"])), NotNil)
	c.Assert(fs.trie.Get(trie.Prefix(fileMap["f3"])), NotNil)
}

func (suite *TheSuite) TestTransientFilesystemer_PathExists(c *C) {
	suite.mkDirs(c, suite.tmpdir)
	fs := suite.mkTFS()
	c.Assert(fs.PathExists("d1"), Equals, true)
	c.Assert(fs.PathExists("/blah"), Equals, false)
}

func (suite *TheSuite) TestTransientFilesystemer_MkPath(c *C) {
	fileMap := suite.mkDirs(c, suite.tmpdir)
	fs := suite.mkTFS()
	err := fs.MkPath("")
	c.Assert(err, IsNil)
	suite.TrieLog(c, fs)
	err = fs.MkPath(filepath.Base(fileMap["f1"]))
	c.Assert(err, Equals, ErrFileExists)
	err = fs.MkPath("d1")
	c.Assert(err, IsNil)
	err = fs.MkPath("d2")
	c.Assert(err, IsNil)
}

func (suite *TheSuite) TestTransientFilesystemer_RmPath(c *C) {
	fileMap := suite.mkDirs(c, suite.tmpdir)
	fs := suite.mkTFS()
	err := fs.RmPath(filepath.Base(fileMap["f1"]))
	c.Assert(err, Equals, ErrFileExists)
	err = fs.RmPath("d1")
	c.Assert(err, IsNil)
	_, err = os.Stat(fileMap["d1"])
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (suite *TheSuite) TestTransientFilesystemer_MkFile(c *C) {
	fs := suite.mkTFS()
	err := fs.MkFile("this/path/here", []byte{})
	c.Assert(err, Equals, ErrPathNoExist) //parent path does not exist
	err = fs.MkFile("here", []byte{})
	c.Assert(err, IsNil)

	cwd, _ := os.Getwd()
	fileBytes, _ := ioutil.ReadFile(filepath.Join(cwd, "thetestfile.txt"))
	err = fs.MkPath("myDir")
	suite.TrieLog(c, fs)
	c.Assert(err, IsNil)
	err = fs.MkFile("myDir/theFile", fileBytes)
	c.Assert(err, IsNil)
	absPath := realPath(fs, "myDir/theFile")
	newBytes, _ := ioutil.ReadFile(absPath)
	c.Assert(fileBytes, DeepEquals, newBytes)

	err = fs.MkFile("myDir/theFile", fileBytes)
	c.Assert(err, Equals, ErrFileExists) //file already exists
}

func (suite *TheSuite) TestTransientFilesystemer_GetFile(c *C) {
	fs := suite.mkTFS()
	_, err := fs.GetFile("dat/path") //file doesn't exist
	c.Assert(err, Equals, ErrFileNoExist)
	err = fs.MkPath("ab/cd")
	_, err = fs.GetFile("ab/cd") //actually a dir
	c.Assert(err, NotNil)

	cwd, _ := os.Getwd()
	fileBytes, _ := ioutil.ReadFile(filepath.Join(cwd, "thetestfile.txt"))
	fs.MkPath("myDir")
	fs.MkFile("myDir/theFile", fileBytes)
	c.Log("MADE IT HERE")

	newFile, err := fs.GetFile("myDir/theFile")
	c.Assert(err, IsNil)
	c.Assert(fileBytes, DeepEquals, newFile.File())

}

package filesystem

import (
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
func (suite *TheSuite) mkFS() *PassThroughFilesystem {
	fs, _ := NewPassThroughFilesystem(suite.tmpdir)
	return fs
}

// tests
func (suite *TheSuite) TestPassThroughFilesystemerBadPath(c *C) {
	_, err := NewPassThroughFilesystem("/thispathbetternotexist")
	c.Assert(err, ErrorMatches, "Unable to stat path.*")
}

func (suite *TheSuite) TestPassThroughFilesystemerFileNotDir(c *C) {
	f, err := ioutil.TempFile(suite.tmpdir, "blah")
	if err != nil {
		c.Fatalf("Failed to make temp file in dir %s", suite.tmpdir)
	}
	_, err = NewPassThroughFilesystem(f.Name())
	c.Assert(err, ErrorMatches, ".*path.*is not a directory")
}

func (suite *TheSuite) TestPassThroughFilesystemerCreate(c *C) {
	fs := suite.mkFS()
	c.Assert(fs.RootPath, Equals, suite.tmpdir)
}

func (suite *TheSuite) TestPassThroughFilesystemer_realPath(c *C) {
	fs := suite.mkFS()
	f, _ := ioutil.TempFile(fs.RootPath, "apples")
	f.Close()
	tmpfileName := filepath.Base(f.Name())
	absPath := fs.realPath(tmpfileName)
	c.Assert(absPath, Equals, filepath.Clean(filepath.Join(fs.RootPath, tmpfileName)))
}

func (suite *TheSuite) TestPassThroughFilesystemer_PathExists(c *C) {
	fs := suite.mkFS()
	f, _ := ioutil.TempFile(fs.RootPath, "apples")
	f.Close()
	tmpfileName := filepath.Base(f.Name())
	c.Assert(fs.PathExists(tmpfileName), Equals, true)
	c.Assert(fs.PathExists("applesssss"), Equals, false)
}

func (suite *TheSuite) TestPassThroughFilesystemer_MkPath(c *C) {
	fs := suite.mkFS()
	err := fs.MkPath("ab/cd")
	c.Assert(err, IsNil)
	c.Assert(fs.PathExists("ab/cd"), Equals, true)
	fs = suite.mkFS()
	err = fs.MkPath("/ab/cd")
	c.Assert(fs.PathExists("/ab/cd"), Equals, true)
	c.Assert(err, IsNil)
	err = fs.MkPath("")
	c.Assert(err, IsNil)
}

func (suite *TheSuite) TestPassThroughFilesystemer_RmPath(c *C) {
	fs := suite.mkFS()
	fs.MkPath("abc/123")
	c.Assert(fs.PathExists("abc/123"), Equals, true)
	err := fs.RmPath("abc/123")
	c.Assert(err, IsNil)
	err = fs.RmPath("abc/notexist")
	c.Assert(err, NotNil)
	err = fs.RmPath("")
	c.Assert(err, NotNil)
	err = fs.RmPath("/")
	c.Assert(err, NotNil)

}

func (suite *TheSuite) TestPassThroughFilesystemer_MkFile(c *C) {
	fs := suite.mkFS()
	err := fs.MkFile("this/path/here", []byte{})
	c.Assert(err, NotNil) //parent path does not exist
	err = fs.MkFile("here", []byte{})
	c.Assert(err, IsNil)

	cwd, _ := os.Getwd()
	fileBytes, _ := ioutil.ReadFile(filepath.Join(cwd, "thetestfile.txt"))
	fs.MkPath("myDir")
	err = fs.MkFile("myDir/theFile", fileBytes)
	c.Assert(err, IsNil)
	absPath := fs.realPath("myDir/theFile")
	newBytes, _ := ioutil.ReadFile(absPath)
	c.Assert(fileBytes, DeepEquals, newBytes)

	err = fs.MkFile("myDir/theFile", fileBytes)
	c.Assert(err, NotNil) //file already exists
}

func (suite *TheSuite) TestPassThroughFilesystemer_GetFile(c *C) {
	fs := suite.mkFS()
	_, err := fs.GetFile("dat/path") //file doesn't exist
	c.Assert(err, NotNil)
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

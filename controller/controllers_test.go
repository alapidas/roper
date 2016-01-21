package controller

import (
	"github.com/alapidas/roper/model"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type TheSuite struct {
	dbpath   string
	repoPath string
	rc       *RoperController
}

var _ = Suite(&TheSuite{})

// Create a temporary directory + db + persister object to use
func (suite *TheSuite) SetUpTest(c *C) {
	tmpdir := c.MkDir()
	tempf, err := ioutil.TempFile(tmpdir, "")
	c.Assert(err, IsNil)
	suite.dbpath = tempf.Name()
	rc, err := Init(suite.dbpath)
	c.Assert(err, IsNil)
	suite.rc = rc
	suite.repoPath = c.MkDir()
}

func (suite *TheSuite) TestBasicPersistAndGetRepo(c *C) {
	repo := &model.Repo{Name: "TestRepo", AbsPath: suite.repoPath}
	pkgs := make(map[string]*model.Package)
	p1_path := "a/b/c.rpm"
	p1 := &model.Package{RepoName: "TestRepo", RelPath: p1_path}
	err := os.MkdirAll(
		filepath.Join(suite.repoPath, filepath.Dir(p1_path)),
		0700,
	)
	c.Assert(err, IsNil)
	err = ioutil.WriteFile(filepath.Join(suite.repoPath, p1_path), []byte(""), 0600)
	c.Assert(err, IsNil)
	pkgs[p1_path] = p1
	repo.Packages = pkgs
	err = suite.rc.PersistRepo(repo)
	c.Assert(err, IsNil)
	pr := &model.PersistableRepo{*repo}
	k,v,err := pr.Serial()
	c.Log("%#v", k)
	c.Log("%#v", v)


	// now get the repo
	repo = nil
	repo, err = suite.rc.GetRepo("TestRepo")
	c.Assert(err, IsNil)
	c.Logf("%#v", repo)

}

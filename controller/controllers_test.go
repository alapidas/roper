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
	dbpath    string
	repoPath  string
	repoPath2 string
	rc        *RoperController
}

var _ = Suite(&TheSuite{})

func (suite *TheSuite) mkPkg(pkg_path, repoName string) (*model.Package, error) {
	p := &model.Package{RepoName: repoName, RelPath: pkg_path}
	if err := os.MkdirAll(
		filepath.Join(suite.repoPath, filepath.Dir(pkg_path)),
		0700,
	); err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(filepath.Join(suite.repoPath, pkg_path), []byte(""), 0600); err != nil {
		return nil, err
	}
	return p, nil
}

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
	suite.repoPath2 = c.MkDir()
}

func (suite *TheSuite) TestBasicPersistAndGetRepo(c *C) {
	repo := &model.Repo{Name: "TestRepo", AbsPath: suite.repoPath}
	pkgs := make(map[string]*model.Package)
	pkg, err := suite.mkPkg("a/b/c.rpm", "TestRepo")
	c.Assert(err, IsNil)
	pkgs[pkg.RelPath] = pkg
	repo.Packages = pkgs

	// persist the repo
	err = suite.rc.PersistRepo(repo)
	c.Assert(err, IsNil)

	// now get the repo
	pstd_repo, err := suite.rc.GetRepo("TestRepo")
	c.Assert(err, IsNil)
	c.Assert(pstd_repo, DeepEquals, repo)

	// persist it again, see what happens
	err = suite.rc.PersistRepo(repo)
	c.Assert(err, IsNil)

	// get it again, make sure it's the same
	pstd_repo, err = suite.rc.GetRepo("TestRepo")
	c.Assert(err, IsNil)
	c.Assert(pstd_repo, DeepEquals, repo)

	// now make the repo have a totally diff package and try again
	newRepo := &model.Repo{Name: "TestRepo", AbsPath: suite.repoPath}
	pkgs = make(map[string]*model.Package)
	pkg, err = suite.mkPkg("d/e.rpm", "TestRepo")
	c.Assert(err, IsNil)
	pkgs[pkg.RelPath] = pkg
	newRepo.Packages = pkgs

	// persist the repo
	err = suite.rc.PersistRepo(newRepo)
	c.Assert(err, IsNil)

	// now get the repo
	pstd_repo, err = suite.rc.GetRepo("TestRepo")
	c.Assert(err, IsNil)
	c.Assert(pstd_repo, DeepEquals, newRepo)
}

func (suite *TheSuite) TestGetRepos(c *C) {
	// make a couple of repos
	repo := &model.Repo{Name: "TestRepo", AbsPath: suite.repoPath}
	pkgs := make(map[string]*model.Package)
	pkg, err := suite.mkPkg("a/b/c.rpm", "TestRepo")
	c.Assert(err, IsNil)
	pkgs[pkg.RelPath] = pkg
	repo.Packages = pkgs

	repo2 := &model.Repo{Name: "TTTT", AbsPath: suite.repoPath2}
	pkgs2 := make(map[string]*model.Package)
	pkg2, err := suite.mkPkg("d/e.rpm", "TTTT")
	c.Assert(err, IsNil)
	pkgs2[pkg2.RelPath] = pkg2
	repo2.Packages = pkgs2

	err = suite.rc.PersistRepo(repo)
	c.Assert(err, IsNil)
	err = suite.rc.PersistRepo(repo2)
	c.Assert(err, IsNil)

	//Now get them
	repos, err := suite.rc.GetRepos()
	c.Assert(err, IsNil)
	c.Assert(len(repos), Equals, 2)

	// return order should be deterministic since keys are sorted alphabetically
	c.Assert(repos[0], DeepEquals, repo2)
	c.Assert(repos[1], DeepEquals, repo)
}


func (suite *TheSuite) TestRemoveRepo(c *C) {
	repo := &model.Repo{Name: "TestRepo", AbsPath: suite.repoPath}
	pkgs := make(map[string]*model.Package)
	pkg, err := suite.mkPkg("a/b/c.rpm", "TestRepo")
	c.Assert(err, IsNil)
	pkgs[pkg.RelPath] = pkg
	repo.Packages = pkgs

	err = suite.rc.PersistRepo(repo)
	c.Assert(err, IsNil)

	repos, err := suite.rc.GetRepos()
	c.Assert(err, IsNil)
	c.Assert(len(repos), Equals, 1)
	c.Assert(repos[0], DeepEquals, repo)

	// Now let's delete it
	err = suite.rc.RemoveRepo("TestRepo")
	c.Assert(err, IsNil)
	repos, err = suite.rc.GetRepos()
	c.Assert(err, IsNil)
	c.Assert(len(repos), Equals, 0)
}
package model

import (
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type TheSuite struct {
}

var _ = Suite(&TheSuite{})

type TestValue struct {
	Value string
}

func (suite *TheSuite) SetUpTest(c *C) {}

func (suite *TheSuite) TestPackage(c *C) {
	p1 := Package{RelPath: "a/b/c"}
	c.Assert(p1.IsRPM(), Equals, false)
	p2 := Package{RelPath: "a/b/c.rpm"}
	c.Assert(p2.IsRPM(), Equals, true)
}

func (suite *TheSuite) TestRepo(c *C) {
	repo := &Repo{Name: "AndysRepo", AbsPath: "/a/b/c", Packages: make(map[string]*Package)}
	c.Assert(len(repo.Packages), Equals, 0)
	p1 := &Package{RelPath: "d/e/f"}

	// add a package
	repo.AddPackage(p1)
	c.Assert(len(repo.Packages), Equals, 1)
	pkg, err := repo.GetPackage("d/e/f")
	c.Assert(err, IsNil)
	c.Assert(pkg, Equals, p1)

	// get nonexistent package
	pkg, err = repo.GetPackage("d/e/ffff")
	c.Assert(err, NotNil)

	// remove nonexistent package
	c.Assert(repo.RmPackage("blah"), NotNil)

	// remove package
	c.Assert(repo.RmPackage("d/e/f"), IsNil)
	pkg, err = repo.GetPackage("d/e/f")
	c.Assert(err, NotNil)
}

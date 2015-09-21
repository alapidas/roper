package model

import (
	. "gopkg.in/check.v1"
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

// tests
func (suite *TheSuite) TestNewRepo(c *C) {
	/*
		_, err := NewRepo("", false)
		c.Assert(err, NotNil)
		_, err = NewRepo("jansdjlasd", false)
		c.Assert(err, NotNil)
		repo, err := NewRepo(suite.tmpdir, false)
		c.Assert(err, IsNil)
		c.Assert(repo, NotNil)
	*/
}

func (suite *TheSuite) TestAddPackage(c *C) {

}

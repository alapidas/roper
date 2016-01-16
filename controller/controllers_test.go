package controller

/*
import (
	"encoding/json"
	"github.com/boltdb/bolt"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type TheSuite struct {
	dbpath    string
	persister *BoltPersistence
}

var _ = Suite(&TheSuite{})

type TestValue struct {
	Value string
}

// Create a temporary directory + db + persister object to use
func (suite *TheSuite) SetUpTest(c *C) {
	tmpdir := c.MkDir()
	tempf, err := ioutil.TempFile(tmpdir, "")
	c.Assert(err, IsNil)
	suite.dbpath = tempf.Name()
	bp, err := CreateBoltPersister(suite.dbpath)
	c.Assert(err, IsNil)
	suite.persister = bp
}
*/

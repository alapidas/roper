package persistence

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

// Create a temporary directory db to use
func (suite *TheSuite) SetUpTest(c *C) {
	tmpdir := c.MkDir()
	tempf, err := ioutil.TempFile(tmpdir, "")
	c.Assert(err, IsNil)
	suite.dbpath = tempf.Name()
	bp, err := CreateBoltPersister(suite.dbpath)
	c.Assert(err, IsNil)
	suite.persister = bp
}

func (suite *TheSuite) mkBucket(name string) error {
	if err := suite.persister.createBucket(name); err != nil {
		return err
	}
	return nil
}

func (suite *TheSuite) TestCreateBucket(c *C) {
	err := suite.persister.createBucket("mario")
	c.Assert(err, IsNil)
	suite.persister.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("mario"))
		c.Assert(b, NotNil)
		return nil
	})

	// test creating it again
	err = suite.persister.createBucket("mario")
	c.Assert(err, IsNil)
}

func (suite *TheSuite) TestPersist(c *C) {
	err := suite.mkBucket("people")
	c.Assert(err, IsNil)
	mario, err := json.Marshal(struct{ Name string }{"mario mario"})
	c.Assert(err, IsNil)
	suite.persister.Persist("people", "1", mario)
	c.Assert(suite.persister.Exists("people", "1"), Equals, true)
	bytes, err := suite.persister.Get("people", "1")
	c.Assert(err, IsNil)
	s := struct{ Name string }{}
	c.Log(bytes)
	err = json.Unmarshal(bytes, &s)
	c.Assert(err, IsNil)
	c.Assert(s, Equals, struct{ Name string }{"mario mario"})
}

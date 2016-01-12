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

func (suite *TheSuite) TestBasicPersist(c *C) {
	err := suite.persister.createBucket("people")
	c.Assert(err, IsNil)
	value := &TestValue{"mario mario"}
	mario := &PersistableBoltItem{"people", "1", *value}
	err = suite.persister.Persist(mario)
	c.Assert(err, IsNil)
	c.Assert(suite.persister.Exists("people", "1"), Equals, true)
	bytes, err := suite.persister.Get("people", "1")
	c.Assert(err, IsNil)
	s := &TestValue{}
	c.Log(bytes)
	err = json.Unmarshal(bytes, s)
	c.Assert(err, IsNil)
	c.Assert(*s, Equals, *value)
}

func (suite *TheSuite) TestInitBuckets(c *C) {
	buckets := []string{"apples", "bananas"}
	err := suite.persister.InitBuckets(buckets)
	c.Assert(err, IsNil)
	// see if the buckets exist by deleting them
	err = suite.persister.Update(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			err := tx.DeleteBucket([]byte(bucket))
			c.Assert(err, IsNil)
		}
		return nil
	})
	c.Assert(err, IsNil)
}

func (suite *TheSuite) TestExists(c *C) {
	bucket := "cars"
	err := suite.persister.InitBuckets([]string{bucket})
	c.Assert(err, IsNil)
	s := "135"
	item := &PersistableBoltItem{bucket, "bmw", &s}
	err = suite.persister.Persist(item)
	c.Assert(err, IsNil)
	c.Assert(suite.persister.Exists(bucket, "bmw"), Equals, true)
	c.Assert(suite.persister.Exists(bucket, "bmwwwww"), Equals, false)
}

func (suite *TheSuite) TestDelete(c *C) {
	bucket := "cars"
	err := suite.persister.InitBuckets([]string{bucket})
	c.Assert(err, IsNil)
	s := "135"
	item := &PersistableBoltItem{bucket, "bmw", &s}
	err = suite.persister.Persist(item)
	c.Assert(suite.persister.Delete(bucket, "bmw"), IsNil)
}

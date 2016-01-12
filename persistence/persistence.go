package persistence

import (
	"fmt"
	"github.com/boltdb/bolt"
	"time"
)

type IPersistableItem interface{}

type IPersistableBoltItem interface {
	Bucket() string
	Key() []byte
	Value() []byte
}

type PersistableBoltItem struct {
	bucket string
	key    interface{}
	value  interface{}
}

// Make PersistableBoltItem methods, and make the BoltPersistence use it.

type IBoltPersister interface {
	/* A persistence interface that's tightly coupled to Bolt for now.
	Uses strings for keys, and marshalled JSON for values in all cases.
	Marshalling and unmarshalling of values will _not_ be attempted here.
	This API does not support nested buckets right now.
	Databases need to be closed by the consumer.
	*/

	// Close a bolt database
	CloseDB() error

	// Persist an item in a bucket
	Persist(bucket, key string, value []byte) error

	// Make sure a given key exists in a bucket
	Exists(bucket, key string) bool

	// Get the value from a given key + bucket.
	Get(bucket, key string) ([]byte, error)

	// Delete a key in a bucket
	// Delete(bucket, key string) error
}

type BoltPersistence struct {
	*bolt.DB
}

var _ IBoltPersister = (*BoltPersistence)(nil)

func CreateBoltPersister(databasePath string) (*BoltPersistence, error) {
	db, err := bolt.Open(databasePath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %s", err)
	}
	return &BoltPersistence{db}, nil
}

func (bp *BoltPersistence) createBucket(bucket string) error {
	err := bp.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create bucket: %s", err)
	}
	return nil
}

func (bp *BoltPersistence) CloseDB() error {
	if err := bp.Close(); err != nil {
		return fmt.Errorf("error closing database: %s", err)
	}
	return nil
}

func (bp *BoltPersistence) Persist(bucket, key string, value []byte) error {
	if err := bp.createBucket(bucket); err != nil {
		return err
	}
	err := bp.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if err := b.Put([]byte(key), value); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unaable to persist item: %s", err)
	}
	return nil
}

func (bp *BoltPersistence) Exists(bucket, key string) bool {
	exists := true
	_ = bp.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		val := b.Get([]byte(key))
		if val == nil {
			exists = false
		}
		return nil
	})
	return exists
}

func (bp *BoltPersistence) Get(bucket, key string) ([]byte, error) {
	dst := []byte{}
	err := bp.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		val := b.Get([]byte(key))
		dst = make([]byte, len(val))
		copy(dst, val)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get item: %s", err)
	}
	return dst, nil
}

package persistence

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"time"
)

type IPersistableItem interface{}

type IPersistableBoltItem interface {
	Bucket() string
	Key() string
	Value() ([]byte, error)
}

type PersistableBoltItem struct {
	bucket string
	key    string
	value  interface{} // Should be a JSON-marshallable item
}

var _ IPersistableBoltItem = (*PersistableBoltItem)(nil)

func (pbi *PersistableBoltItem) Bucket() string { return pbi.bucket }
func (pbi *PersistableBoltItem) Key() string    { return pbi.key } // TODO: Convert to string
func (pbi *PersistableBoltItem) Value() ([]byte, error) {
	bytes, err := json.Marshal(pbi.value)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal value: %s", err)
	}
	return bytes, nil
}

type IBoltPersister interface {
	/* A persistence interface that's tightly coupled to Bolt for now.
	Marshalling and unmarshalling of values will _not_ be attempted in this level.
	This API does not support nested buckets right now.
	Databases need to be closed by the consumer.
	Top level buckets need to explicitly be initialized via InitBuckets() before use
	*/

	// Close a bolt database
	CloseDB() error

	// Initialize buckets
	InitBuckets(buckets []string) error

	// Persist an item in a bucket
	Persist(item IPersistableBoltItem) error

	// Make sure a given key exists in a bucket
	Exists(bucket, key string) bool

	// Get the value from a given key + bucket.
	Get(bucket, key string) ([]byte, error)

	// Delete a key in a bucket
	Delete(bucket, key string) error
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
	bp := &BoltPersistence{db}
	return bp, nil
}

func (bp *BoltPersistence) InitBuckets(buckets []string) error {
	err := bp.Update(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create bucket: %s", err)
	}
	return nil
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

func (bp *BoltPersistence) Persist(item IPersistableBoltItem) error {
	// First ensure we can get a value from the item
	value, err := item.Value()
	if err != nil {
		return fmt.Errorf("unable to persist item value: %s", err)
	}
	err = bp.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(item.Bucket()))
		if err := b.Put([]byte(item.Key()), value); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to persist item: %s", err)
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

func (bp *BoltPersistence) Delete(bucket, key string) error {
	err := bp.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if err := b.Delete([]byte(key)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to delete item as key %s: %s", key, err)
	}
	return nil
}

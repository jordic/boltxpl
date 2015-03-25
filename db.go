package boltxpl

import (
	"os"
	"time"

	"strings"

	"github.com/boltdb/bolt"
)

type DB struct {
	*bolt.DB
}

func (db *DB) Open(path string, mode os.FileMode) error {
	d, err := bolt.Open(path, mode, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	db.DB = d
	return nil
}

func (db *DB) View(fn func(*Tx) error) error {
	return db.DB.View(func(tx *bolt.Tx) error {
		return fn(&Tx{tx})
	})
}

// Update executes a function in the context of a writable transaction.
func (db *DB) Update(fn func(*Tx) error) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		return fn(&Tx{tx})
	})
}

type Tx struct {
	*bolt.Tx
}

func (tx *Tx) NestedBucket(b string) *bolt.Bucket {
	keys := strings.Split(b, "/")
	var bucket *bolt.Bucket
	for i, k := range keys {
		if i == 0 {
			bucket = tx.Bucket([]byte(k))
		} else {
			bucket = bucket.Bucket([]byte(k))
		}
		if bucket == nil {
			break
		}
	}

	return bucket
}

type DBItem struct {
	Key      string
	IsBucket bool
}
package boltxpl

import (
	"os"
	"testing"
)

func TestNestedBucketsGet(t *testing.T) {
	f := tempfile()
	ddb := OpenDb(f)
	defer ddb.Close()

	ddb.Update(func(tx *Tx) error {
		tx.CreateBucketIfNotExists([]byte("test1"))
		b := tx.Bucket([]byte("test1"))
		b.CreateBucketIfNotExists([]byte("test2"))
		b.CreateBucketIfNotExists([]byte("test3"))
		c := b.Bucket([]byte("test3"))
		c.CreateBucketIfNotExists([]byte("test4"))
		d := c.Bucket([]byte("test4"))
		d.Put([]byte("t"), []byte("1"))
		return nil
	})

	ddb.View(func(tx *Tx) error {
		b := tx.NestedBucket("test1/test2")
		if b == nil {
			t.Error("must return a bucket")
		}
		b1 := tx.NestedBucket("test1/test3")
		if b1 == nil {
			t.Error("must return a bucket")
		}
		end := b1.Bucket([]byte("test4"))
		if end == nil {
			t.Error("must return a bucket")
		}

		x := end.Get([]byte("t"))
		if string(x) != "1" {
			t.Error("must return a bucket")
		}

		xx := tx.NestedBucket("test1/test3/test4")
		x = xx.Get([]byte("t"))
		if string(x) != "1" {
			t.Error("must return a bucket")
		}
		
		xx = tx.NestedBucket("test1/test3/test4/t")
		if xx != nil {
			t.Error("must return a bucket")
		}
		
		
		
		return nil
	})

	os.Remove(f)

}
package boltxpl

import (
	"encoding/json"	
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
)

func tempfile() string {
	f, _ := ioutil.TempFile("", "bolt-")
	f.Close()
	os.Remove(f.Name())
	return f.Name()
}

func OpenDb(f string) DB {

	var db DB
	if err := db.Open(f, 0600); err != nil {
		log.Fatal("db:", err)
	}
	return db

}

func TestGetBucket(t *testing.T) {
	f := tempfile()
	ddb := OpenDb(f)
	defer ddb.Close()
	h := NewHandler(&ddb)
	ddb.Update(func(tx *Tx) error {
		tx.CreateBucketIfNotExists([]byte("test1"))
		tx.CreateBucketIfNotExists([]byte("test2"))
		c := tx.Bucket([]byte("test1"))
		c.CreateBucketIfNotExists([]byte("test3"))
		d := c.Bucket([]byte("test3"))
		d.Put([]byte("a"), []byte("1"))
		d.Put([]byte("b"), []byte("2"))
		d.CreateBucketIfNotExists([]byte("c"))
		return nil
	})
	
	r, _ := http.NewRequest("GET", "/bucket/test1/test3", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Error("Request should return Ok 200")
	}
	res := make([]DBItem, 0)
	decoder := json.NewDecoder(w.Body)
	_ = decoder.Decode(&res)
	if len(res) != 3 {
		t.Error("Expected %s got %s", 3, len(res))
	}
	if res[0].Key != "a" {
		t.Error("Expected %s got %s", "a", res[0].Key)
	}
	if res[2].Key != "c" {
		t.Error("Expected %s got %s", "c", res[2].Key)
	}
	if res[2].IsBucket != true {
		t.Error("key 2 should be a bucket")
	}

}

func TestRootHandler(t *testing.T) {
	f := tempfile()
	ddb := OpenDb(f)
	defer ddb.Close()
	h := NewHandler(&ddb)

	ddb.Update(func(tx *Tx) error {
		tx.CreateBucketIfNotExists([]byte("test1"))
		tx.CreateBucketIfNotExists([]byte("test2"))
		return nil
	})

	r, err := http.NewRequest("GET", "/root", nil)
	if err != nil {
		t.Errorf("ERR:%s", err)
	}
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Error("Request should return Ok 200")
	}

	res := make([]DBItem, 0)
	decoder := json.NewDecoder(w.Body)
	err = decoder.Decode(&res)

	if len(res) != 2 {
		t.Error("Expected json response")
	}

	if res[0].Key != "test1" {
		t.Error("Wrong key for first item")
	}

	if res[0].IsBucket != true {
		t.Error("Bucket should be true")
	}

	os.Remove(f)

}

func TestMuxRoutes(t *testing.T) {

	m := mux.NewRouter()
	m.HandleFunc("/test/{key:.*}", func(w http.ResponseWriter, r *http.Request) {
		item := mux.Vars(r)
		r.Header.Set("Path", item["key"])
	}).Methods("GET")

	r, err := http.NewRequest("GET", "/test/111/222/333", nil)
	if err != nil {
		t.Errorf("ERR:%s", err)
	}
	w := httptest.NewRecorder()
	m.ServeHTTP(w, r)

	if r.Header.Get("Path") != "111/222/333" {
		t.Errorf("Expected %s, got %s", "111/222/333", r.Header["Path"])
	}

}
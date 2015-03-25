package boltxpl

import (
	"encoding/json"
	

	"net/http"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)



func NewHandler(db *DB) http.Handler {

	h := &Handler{DB: db}

	r := mux.NewRouter()
	r.HandleFunc("/root", h.Root).Methods("GET")
	r.HandleFunc("/bucket/{bucket:.*}", h.GetBucket).Methods("GET")

	return r
}


type Handler struct {
	DB *DB
}

func (h *Handler) Root(w http.ResponseWriter, req *http.Request) {
	buckets := make([]DBItem, 0)
	_ = h.DB.View(func(tx *Tx) error {
		tx.ForEach(func(k []byte, bucket *bolt.Bucket) error {
			item := DBItem{
				Key:      string(k),
				IsBucket: true,
			}
			buckets = append(buckets, item)
			return nil
		})
		return nil
	})

	enc, _ := json.Marshal(buckets)
	w.Header().Set("Content-Type", "application/json")
	w.Write(enc)
}



func (h *Handler) GetBucket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["bucket"]
	key = strings.TrimSuffix(key, "/")
	
	result := make([]DBItem, 0)
	_ = h.DB.View(func(tx *Tx) error {
		
		b := tx.NestedBucket(key)
		if b == nil {
			return nil
		}
		b.ForEach(func(k, v []byte) error {
			s := DBItem{
				Key: string(k),
			}
			if v == nil {
				s.IsBucket = true
			}
			result = append(result, s)
			return nil
		})		
		return nil
	})
	
	enc, _:= json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(enc)	
}



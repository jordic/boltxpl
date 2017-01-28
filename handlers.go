package boltxpl

import (
	"encoding/binary"
	"encoding/json"
	"unicode/utf8"

	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

const VERSION = "0.2"

func NewHandler(db *DB) http.Handler {

	h := &Handler{
		DB: db,
	}

	r := mux.NewRouter()
	r.HandleFunc("/", h.Index).Methods("GET")
	r.HandleFunc("/root", h.Root).Methods("GET")
	r.HandleFunc("/bucket/{bucket:.*}", h.GetBucket).Methods("GET")
	r.HandleFunc("/-view/", h.ViewKey).Methods("GET")
	return r
}

type Handler struct {
	DB *DB
}

func (h *Handler) Root(w http.ResponseWriter, req *http.Request) {
	buckets := make([]DBItem, 0)
	_ = h.DB.View(func(tx *Tx) error {
		tx.ForEach(func(k []byte, bucket *bolt.Bucket) error {

			//check for invalid string
			keyString := string(k)
			if !utf8.ValidString(keyString) {
				keyString = strconv.Itoa(int(binary.BigEndian.Uint64(k)))
			}

			item := DBItem{
				Key:      keyString,
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
	r.ParseForm()
	seek := r.Form.Get("p")
	perpage := 15
	cur := 0

	result := make([]DBItem, 0)
	_ = h.DB.View(func(tx *Tx) error {

		b := tx.NestedBucket(key)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		if seek == "" {
			key, _ := c.First()
			seek = string(key)
		}

		for k, v := c.Seek([]byte(seek)); k != nil; k, v = c.Next() {

			//check for invalid string
			keyString := string(k)
			if !utf8.ValidString(keyString) {
				keyString = strconv.Itoa(int(binary.BigEndian.Uint64(k)))
			}

			s := DBItem{
				Key: keyString,
			}

			if v == nil {
				s.IsBucket = true
			}
			result = append(result, s)
			cur++
			if cur > perpage {
				break
			}
		}
		return nil
	})

	enc, _ := json.Marshal(result)
	fmt.Println("sending keys for bucket", string(enc))
	w.Header().Set("Content-Type", "application/json")
	w.Write(enc)
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("template.html").Parse(TemplateHTML)

	//t, err := template.New("listing").ParseFiles("cmd/xpl/template.html")
	if err != nil {
		log.Fatalf("Template error %s", err)
	}
	v := map[string]interface{}{
		"version": VERSION,
	}

	w.Header().Set("Content-Type", "text/html")
	err = t.ExecuteTemplate(w, "template.html", v)
	if err != nil {
		log.Fatalf("Template error %s", err)
	}
}

func (h *Handler) ViewKey(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	bucket := r.Form.Get("bucket")
	key := r.Form.Get("key")
	fmt.Println("Got viewkey key:", key)
	var res []byte
	_ = h.DB.View(func(tx *Tx) error {
		b := tx.NestedBucket(bucket)
		if b == nil {
			return nil
		}
		res = b.Get([]byte(key))
		// check for nil and try as int(64)
		if res == nil {
			i, _ := strconv.Atoi(key)
			res = b.Get(itob(int64(i)))
		}
		return nil
	})
	w.Write(res)

}

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

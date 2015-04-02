package main


import (
	"flag"
	"log"
	"net/http"
	"github.com/jordic/boltxpl"
)



func main() {

	file := flag.String("db", "test.db", "database file")
	addr := flag.String("address", ":5000", "bind adress")
	
	flag.Parse()
	log.SetFlags(0)
	
	var db boltxpl.DB
	if err := db.Open(*file, 0600); err != nil {
		log.Fatal("db:", err)
	}

	log.Printf("Listening on http://%s", *addr)
	
	
	
	log.Fatal( http.ListenAndServe(*addr, boltxpl.NewHandler(&db)) )

}
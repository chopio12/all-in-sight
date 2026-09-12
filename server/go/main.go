package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

const (
	dbPath = "../poker-go.db"
	addr   = ":8001"
)

func main() {
	fmt.Printf("Starting server on %s\n", addr)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := initDB(db); err != nil {
		log.Fatal(err)
	}

	app := &server{db: db}
	mux := newRouter(app)

	log.Printf("Poker API listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
)

var db *sql.DB

func GetBrands(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select distinct brand from inks order by brand asc")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var brands []string
	for rows.Next() {
		var brand string
		err := rows.Scan(&brand)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		brands = append(brands, brand)
	}

	json.NewEncoder(w).Encode(brands)
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./db.sql")
	if err != nil {
		log.Fatalf("Could not open SQL database: %s\n", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/brands", GetBrands)

	address := os.Getenv("API_ADDR")
	if address == "" {
		address = "localhost:4000"
	}

	log.Printf("Listening on [%s]\n", address)
	log.Fatal(http.ListenAndServe(address, mux))
}

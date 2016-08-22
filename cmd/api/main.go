package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"time"
)

var db *sql.DB

type Brand struct {
	Name, Slug string
}

type Ink struct {
	Brand      Brand
	Name, Slug string
}

type InkDetailed struct {
	Brand                  Brand
	Name, Slug, URL, Donor string
}

func GetBrands(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select distinct brand, brand_slug from inks order by brand asc")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var brands []Brand
	for rows.Next() {
		var brand Brand
		err := rows.Scan(&brand.Name, &brand.Slug)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		brands = append(brands, brand)
	}

	json.NewEncoder(w).Encode(brands)
}

func GetBrandInks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	brandSlug := vars["brandSlug"]

	rows, err := db.Query("select brand, brand_slug, name, name_slug from inks where brand_slug = ? order by name asc", brandSlug)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var (
		inks  []Ink
		brand Brand
		found bool = false
	)
	for rows.Next() {
		found = true
		var ink Ink
		err := rows.Scan(&brand.Name, &brand.Slug, &ink.Name, &ink.Slug)
		ink.Brand = brand

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		inks = append(inks, ink)
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "[]")
		return
	}

	json.NewEncoder(w).Encode(inks)
}

func GetInk(w http.ResponseWriter, r *http.Request) {
	var (
		ink   InkDetailed
		brand Brand
		vars  map[string]string = mux.Vars(r)
	)

	brand.Slug, ink.Slug = vars["brandSlug"], vars["inkSlug"]

	err := db.QueryRow("select brand, name, url, donor from inks where brand_slug = ? and name_slug = ?", brand.Slug, ink.Slug).Scan(&brand.Name, &ink.Name, &ink.URL, &ink.Donor)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	ink.Brand = brand

	json.NewEncoder(w).Encode(ink)
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./db.sql")
	if err != nil {
		log.Fatalf("Could not open SQL database: %s\n", err)
	}
	defer db.Close()

	r := mux.NewRouter().PathPrefix("/v1/").Subrouter()
	r.HandleFunc("/brands", GetBrands).Methods("GET")
	r.HandleFunc("/brands/{brandSlug}/inks", GetBrandInks).Methods("GET")
	r.HandleFunc("/brands/{brandSlug}/inks/{inkSlug}", GetInk).Methods("GET")

	address := os.Getenv("API_ADDR")
	if address == "" {
		address = "localhost:4000"
	}

	log.Printf("Listening on [%s]\n", address)
	srv := &http.Server{
		Handler: r,
		Addr:    address,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

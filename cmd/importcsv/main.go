package main

import (
	"database/sql"
	"encoding/csv"
	"github.com/kamaln7/swabs"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"os"
)

func CreateTable(tx *sql.Tx) {
	schema := `
	CREATE TABLE IF NOT EXISTS inks (
		id INTEGER PRIMARY KEY,
		name TEXT UNIQUE,
		url TEXT,
		donor TEXT
	)`

	_, err := tx.Exec(schema)
	if err != nil {
		log.Fatalf("Could not initialize SQL database: %s\n", err)
	}
}

func main() {
	db, err := sql.Open("sqlite3", "./db.sql")
	defer db.Close()

	file, err := os.Open("./data.csv")
	if err != nil {
		log.Fatalf("Could not open SQL database: %s\n", err)
	}
	defer file.Close()

	var cols []string = nil
	log.Print("Starting SQL transaction")
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("An SQL error occurred: %s\n", err)
	}
	defer tx.Commit()

	CreateTable(tx)

	insert, err := tx.Prepare("insert into swabs(name, url, donor) values (?, ?, ?)")
	if err != nil {
		log.Fatalf("An error occured while preparing the SQL query: %s\n", err)
	}

	reader := csv.NewReader(file)
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				log.Print("Reached end of file\n")
				break
			}

			if perr, ok := err.(*csv.ParseError); ok {
				log.Printf("Could not parse line %d column %d: %s\n", perr.Line, perr.Column, perr.Err)
				continue
			}

			// unknown error, probably best to exit
			log.Print("Rolling back SQL transaction\n")
			tx.Rollback()
			log.Fatalf("An unexpected error occured: %s\n", err)
		}

		if cols == nil {
			cols = row
			continue
		}

		var swab swabs.Ink
		for i, v := range row {
			switch cols[i] {
			case "Name":
				swab.Name = v
			case "Imgur Address":
				swab.URL = v
			case "Donated by":
				swab.Donor = v
			}

		}

		log.Printf("Importing [%s]\n", swab.Name)
		_, err = insert.Exec(swab.Name, swab.URL, swab.Donor)
		if err != nil {
			log.Printf("Error importing [%s]: %s\n", swab.Name, err)
			continue
		}
	}
}

package main

import (
	"database/sql"
	"encoding/csv"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"os"
)

type Ink struct {
	name, url, donor string
}

func FatalOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func CreateTable(tx *sql.Tx) {
	schema := `
	CREATE TABLE IF NOT EXISTS swabs (
		id INTEGER PRIMARY KEY,
		name TEXT UNIQUE,
		url TEXT,
		donor TEXT
	)`

	_, err := tx.Exec(schema)
	FatalOnErr(err)
}

func main() {
	db, err := sql.Open("sqlite3", "./db.sql")
	defer db.Close()

	file, err := os.Open("./data.csv")
	FatalOnErr(err)
	defer file.Close()

	var cols []string = nil
	log.Print("Starting SQL transaction")
	tx, err := db.Begin()
	FatalOnErr(err)
	defer tx.Commit()

	CreateTable(tx)

	insert, err := tx.Prepare("insert into swabs(name, url, donor) values (?, ?, ?)")
	FatalOnErr(err)

	reader := csv.NewReader(file)
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Print(err)
			continue
		}

		if cols == nil {
			cols = row
		} else {
			var swab Ink
			for i, v := range row {
				switch cols[i] {
				case "Name":
					swab.name = v
				case "Imgur Address":
					swab.url = v
				case "Donated by":
					swab.donor = v
				}

			}

			log.Printf("--> Importing [%s]\n", swab.name)
			_, err = insert.Exec(swab.name, swab.url, swab.donor)
			if err != nil {
				log.Printf("<-- Error importing [%s]\n", swab.name)
				log.Printf("<-- %s\n", err.Error())
				continue
			}
		}
	}
}

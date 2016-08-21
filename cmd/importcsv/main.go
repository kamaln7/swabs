package main

import (
	"database/sql"
	"encoding/csv"
	"github.com/Machiel/slugify"
	"github.com/kamaln7/swabs"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"os"
	"strings"
)

func CreateTable(tx *sql.Tx) {
	schema := `
	CREATE TABLE IF NOT EXISTS inks (
		id INTEGER PRIMARY KEY,
		brand TEXT not null,
		brand_slug TEXT not null,
		name TEXT not null,
		name_slug TEXT not null,
		url TEXT not null,
		donor TEXT
	)
	`

	_, err := tx.Exec(schema)
	if err != nil {
		log.Fatalf("Could not initialize SQL database: %s\n", err)
	}

	indexes := `
	CREATE INDEX IF NOT EXISTS idx_ink_brand ON inks(brand);
	CREATE INDEX IF NOT EXISTS idx_ink_name ON inks(name);
	CREATE UNIQUE INDEX uq_ink_brand_name ON inks(brand, name);
	`

	_, err = tx.Exec(indexes)
	if err != nil {
		log.Fatalf("Could not create indexes: %s\n", err)
	}
}

type Brand struct {
	Names []string
}

func ReadBrands() []Brand {
	brands := make([]Brand, 0)

	file, err := os.Open("./data/brands.csv")
	defer file.Close()

	if err != nil {
		log.Fatalf("Could not open brands.csv: %s\n", err)
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // not all lines have an equal amount of fields
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
			log.Fatalf("An unexpected error occured: %s\n", err)
		}

		var brand Brand
		for _, name := range row {
			brand.Names = append(brand.Names, name)
		}

		brands = append(brands, brand)
		log.Printf("Read brand [%s]\n", brand.Names[0])
	}

	return brands
}

func main() {
	db, err := sql.Open("sqlite3", "./db.sql")
	if err != nil {
		log.Fatalf("Could not open SQL database: %s\n", err)
	}
	defer db.Close()

	brands := ReadBrands()

	inks, err := os.Open("./data/inks.csv")
	if err != nil {
		log.Fatalf("Could not open inks.csv: %s\n", err)
	}
	defer inks.Close()

	var cols []string = nil
	log.Print("Starting SQL transaction")
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("An SQL error occurred: %s\n", err)
	}
	defer tx.Commit()

	CreateTable(tx)

	insert, err := tx.Prepare("insert into inks(brand, brand_slug, name, name_slug, url, donor) values (?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatalf("An error occured while preparing the SQL query: %s\n", err)
	}

	reader := csv.NewReader(inks)
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

		var (
			swab swabs.Ink
		)

		for i, v := range row {
			switch cols[i] {
			case "Name":
				swab.Name = strings.Replace(v, ":", "", -1)
			case "Imgur Address":
				swab.URL = v
			case "Donated by":
				swab.Donor = v
			}
		}

		// forgive me
	GetBrandName:
		for _, brand := range brands {
			for _, brandName := range brand.Names {
				if strings.HasPrefix(strings.ToLower(swab.Name), strings.ToLower(brandName)) {
					swab.Brand = brand.Names[0]
					swab.Name = swab.Name[len(brandName)+1:]
					break GetBrandName
					break
				}
			}
		}

		if swab.Brand == "" {
			log.Printf("Error importing [%s]: could not find matching brand\n", swab.Name)
			continue
		}

		log.Printf("Importing [%s] - [%s]\n", swab.Brand, swab.Name)
		_, err = insert.Exec(
			swab.Brand,
			slugify.Slugify(swab.Brand),
			swab.Name,
			slugify.Slugify(swab.Name),
			swab.URL,
			swab.Donor,
		)
		if err != nil {
			log.Printf("Error importing [%s] - [%s]: %s\n", swab.Brand, swab.Name, err)
			continue
		}
	}
}

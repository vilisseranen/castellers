package model

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/coreos/go-semver/semver"
)

const initQuery = `CREATE TABLE IF NOT EXISTS schema_version
(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	version TEXT NOT NULL,
    installed INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT version UNIQUE (version)
);`

var db *sql.DB

func InitializeDB(dbname string) {
	var err error
	db, err = sql.Open("sqlite3", dbname)
	if err != nil {
		log.Fatal(err)
	}
	// Prepare the schema_version table
	_, err = db.Exec(initQuery)
	if err != nil {
		log.Fatal(err)
	}

	// List all update files
	files, err := ioutil.ReadDir("sql/")
	if err != nil {
		log.Fatal(err)
	}
	vs := make([]*semver.Version, len(files))
	for i, f := range files {
		v, err := semver.NewVersion(f.Name()[:len(f.Name())-4])
		if err != nil {
			log.Fatal(err)
		} else {
			vs[i] = v
		}
	}

	// Sort update files
	semver.Sort(vs)

	// For each, check if it was applied
	for _, v := range vs {
		version := fmt.Sprintf("%s", v)
		if schemaInstalled(version) == false {
			content, err := ioutil.ReadFile(fmt.Sprintf("sql/%s.sql", version))
			if err == nil {
				fmt.Printf("Updating schema to %s\n", version)
				tx, err := db.Begin()
				if err != nil {
					log.Fatal(err)
				}
				_, err = tx.Exec(string(content))
				if err != nil {
					log.Fatal(err)
				}
				stmt, err := tx.Prepare("INSERT INTO schema_version (version, installed) VALUES(?, 1);")
				defer stmt.Close()
				if err != nil {
					log.Fatal(err)
				}
				_, err = stmt.Exec(version)
				if err != nil {
					log.Fatal(err)
				}
				err = tx.Commit()
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
		} else {
			fmt.Printf("Schema %s already installed.\n", version)
		}
	}

}

func stringOrNull(s string) string {
	if len(s) == 0 {
		return ""
	} else {
		return s
	}
}

func schemaInstalled(version string) bool {
	installed := 0
	stmt, err := db.Prepare("SELECT COUNT(id) FROM schema_version WHERE installed = 1 AND version = ?;")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(
		version).Scan(&installed)
	if err != nil {
		log.Fatal(err)
	}
	return installed == 1
}

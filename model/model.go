package model

import (
	"database/sql"
	"fmt"
	"io/ioutil"

	"github.com/coreos/go-semver/semver"
	"github.com/vilisseranen/castellers/common"
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
	db, err = sql.Open("sqlite3", dbname+"?_foreign_keys=on")
	if err != nil {
		common.Fatal(err.Error())
	}
	// Prepare the schema_version table
	_, err = db.Exec(initQuery)
	if err != nil {
		common.Fatal(err.Error())
	}

	// List all update files
	files, err := ioutil.ReadDir("sql/")
	if err != nil {
		common.Fatal(err.Error())
	}
	vs := make([]*semver.Version, len(files))
	for i, f := range files {
		v, err := semver.NewVersion(f.Name()[:len(f.Name())-4])
		if err != nil {
			common.Fatal(err.Error())
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
				common.Info("Updating schema to %s\n", version)
				tx, err := db.Begin()
				if err != nil {
					common.Fatal(err.Error())
				}
				_, err = tx.Exec(string(content))
				if err != nil {
					common.Fatal(err.Error())
				}
				stmt, err := tx.Prepare("INSERT INTO schema_version (version, installed) VALUES(?, 1);")
				defer stmt.Close()
				if err != nil {
					common.Fatal(err.Error())
				}
				_, err = stmt.Exec(version)
				if err != nil {
					common.Fatal(err.Error())
				}
				err = tx.Commit()
				if err != nil {
					common.Fatal(err.Error())
				}
			} else {
				common.Fatal(err.Error())
			}
			common.Info("Update schema to %s completed\n", version)
		} else {
			common.Info("Schema %s already installed.\n", version)
		}
	}

}

func stringOrNull(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func nullToEmptyString(s sql.NullString) string {
	if !s.Valid {
		return ""
	}
	return s.String
}

func schemaInstalled(version string) bool {
	installed := 0
	stmt, err := db.Prepare("SELECT COUNT(id) FROM schema_version WHERE installed = 1 AND version = ?;")
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	err = stmt.QueryRow(
		version).Scan(&installed)
	if err != nil {
		common.Fatal(err.Error())
	}
	return installed == 1
}

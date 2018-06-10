package model

import (
	"database/sql"
	"log"
)

var db *sql.DB

// Tables names

type Entity interface {
	Get() error
	GetAll() error
}

func InitializeDB(dbname string) {
	var err error
	db, err = sql.Open("sqlite3", dbname)
	if err != nil {
		log.Fatal(err)
	}
}

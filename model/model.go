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
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(EventsTableCreationQuery)
	_, err = tx.Exec(MembersTableCreationQuery)
	_, err = tx.Exec(ParticipationTableCreationQuery)
	tx.Commit()
}

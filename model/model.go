package model

import (
	"database/sql"
	"log"
)

var db *sql.DB

// TODO: remove this interface
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
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(MembersTableCreationQuery)
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(ParticipationTableCreationQuery)
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(RecurringEventsTableCreationQuery)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

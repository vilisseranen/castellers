package model

import (
	"database/sql"
	"log"
)

var db *sql.DB

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

// From: https://stackoverflow.com/questions/40266633/golang-insert-null-into-sql-instead-of-empty-string
func stringOrNull(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

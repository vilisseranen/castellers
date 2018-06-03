package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
)

// Tables names
const EVENTS_TABLE = "events"
const MEMBERS_TABLE = "members"
const ADMIN_TABLE = "admins"

// Tables creation queries
const EventsTableCreationQuery = `CREATE TABLE IF NOT EXISTS events
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	startDate TEXT NOT NULL,
	endDate TEXT NOT NULL,
	description TEXT,
	uuid TEXT NOT NULL,
	CONSTRAINT uuid_unique UNIQUE (uuid)
)
;`
const AdminsTableCreationQuery = `CREATE TABLE IF NOT EXISTS admins
( id INTEGER PRIMARY KEY AUTOINCREMENT, uuid TEXT NOT NULL );`

type event struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

func (e *event) getEvent(db *sql.DB) error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT name, startDate, endDate FROM %s WHERE uuid= ?", EVENTS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(e.UUID).Scan(&e.Name, &e.StartDate, &e.EndDate)
	return err
}

func (e *event) getEvents(db *sql.DB, start, count int) ([]event, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT uuid, name, startDate, endDate FROM %s LIMIT ? OFFSET ?", EVENTS_TABLE), count, start)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	events := []event{}

	for rows.Next() {
		var e event
		if err = rows.Scan(&e.UUID, &e.Name, &e.StartDate, &e.EndDate); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (e *event) updateEvent(db *sql.DB) error {
	stmt, err := db.Prepare(fmt.Sprintf("Update %s SET name = ?, startDate = ?, endDate = ? WHERE uuid= ?", EVENTS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(e.Name, e.StartDate, e.EndDate, e.UUID)
	return err
}

func (e *event) deleteEvent(db *sql.DB) error {
	stmt, err := db.Prepare(fmt.Sprintf("DELETE FROM %s WHERE uuid= ?", EVENTS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(e.UUID)
	return err
}

func (e *event) createEvent(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (uuid, name, startDate, endDate) VALUES (?, ?, ?, ?)", EVENTS_TABLE))
	if err != nil {
		return err
	}
	defer stmt.Close()
	e.generateUUID()
	_, err = stmt.Exec(e.UUID, e.Name, e.StartDate, e.EndDate)
	if err != nil {
		return err
	}
	tx.Commit()
	return err
}

func (e *event) generateUUID() {
	data := make([]byte, 10)
	_, err := rand.Read(data)
	if err != nil {
		log.Fatal(err)
	}
	e.UUID = fmt.Sprintf("%x", sha256.Sum256(data))
}

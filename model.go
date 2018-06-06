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
const ADMINS_TABLE = "admins"
const PRESENCES_TABLE = "presences"

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
);`
const AdminsTableCreationQuery = `CREATE TABLE IF NOT EXISTS admins
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uuid TEXT NOT NULL,
	CONSTRAINT uuid_unique UNIQUE (uuid)
);`
const MembersTableCreationQuery = `CREATE TABLE IF NOT EXISTS members
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uuid TEXT NOT NULL,
	name TEXT NOT NULL,
	extra TEXT,
	CONSTRAINT uuid_unique UNIQUE (uuid)
);`

const PresencesTableCreationQuery = `CREATE TABLE IF NOT EXISTS presences
(
	member_id INTEGER NOT NULL,
	event_id INTEGER NOT NULL,
  answer TEXT NOT NULL,
	PRIMARY KEY (member_id, event_id)
);`

const UUID_SIZE = 40

type event struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type admin struct {
	UUID string `json:"uuid"`
}

type member struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Extra string `json:"extra"`
}

func (a *admin) getAdmin(db *sql.DB) error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT uuid FROM %s WHERE uuid= ?", ADMINS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(a.UUID).Scan(&a.UUID)
	return err
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

func (m *member) createMember(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (uuid, name, extra) VALUES (?, ?, ?)", MEMBERS_TABLE))
	if err != nil {
		return err
	}
	defer stmt.Close()
	m.generateUUID()
	_, err = stmt.Exec(m.UUID, m.Name, m.Extra)
	if err != nil {
		return err
	}
	tx.Commit()
	return err
}

func (m *member) getMember(db *sql.DB) error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT name, extra FROM %s WHERE uuid= ?", MEMBERS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(m.UUID).Scan(&m.Name, &m.Extra)
	return err
}

func (e *event) generateUUID() {
	data := make([]byte, 10)
	_, err := rand.Read(data)
	if err != nil {
		log.Fatal(err)
	}
	e.UUID = fmt.Sprintf("%x", sha256.Sum256(data))[:UUID_SIZE]
}

func (m *member) generateUUID() {
	data := make([]byte, 10)
	_, err := rand.Read(data)
	if err != nil {
		log.Fatal(err)
	}
	m.UUID = fmt.Sprintf("%x", sha256.Sum256(data))[:UUID_SIZE]
}

package model

import (
	"fmt"
	"log"
)

const RECURRING_EVENTS_TABLE = "recurring_events"

const RecurringEventsTableCreationQuery = `CREATE TABLE IF NOT EXISTS recurring_events
(
	uuid TEXT PRIMARY KEY,
	name TEXT NOT NULL,
  description TEXT,
	interval TEXT NOT NULL
);`

type RecurringEvent struct {
	UUID        string
	Name        string
	Description string
	Interval    string
}

func (r *RecurringEvent) Get() error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT name, description, interval FROM %s WHERE uuid= ?", RECURRING_EVENTS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(r.UUID).Scan(&r.Name, &r.Description, &r.Interval)
	return err
}

func (r *RecurringEvent) CreateRecurringEvent() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (uuid, name, description, interval) VALUES (?, ?, ?, ?)", RECURRING_EVENTS_TABLE))
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(r.UUID, r.Name, r.Description, r.Interval)
	if err != nil {
		return err
	}
	tx.Commit()
	return err
}

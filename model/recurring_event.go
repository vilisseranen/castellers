package model

import (
	"fmt"

	"github.com/vilisseranen/castellers/common"
)

const RECURRING_EVENTS_TABLE = "recurring_events"

type RecurringEvent struct {
	UUID        string
	Name        string
	Description string
	Interval    string
}

func (r *RecurringEvent) Get() error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT name, description, interval FROM %s WHERE uuid= ?", RECURRING_EVENTS_TABLE))
	if err != nil {
		common.Fatal(err.Error())
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
	err = tx.Commit()
	return err
}

package model

import (
	"fmt"
	"log"
	"time"
)

// TODO: implement deleted flag on events

const EVENTS_TABLE = "events"

const EventsTableCreationQuery = `CREATE TABLE IF NOT EXISTS events
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	startDate INTEGER NOT NULL,
	endDate INTEGER NOT NULL,
	description TEXT,
	uuid TEXT NOT NULL,
	recurringEvent TEXT,
	type TEXT NOT NULL,
	CONSTRAINT uuid_unique UNIQUE (uuid),
	FOREIGN KEY(recurringEvent) REFERENCES recurring_events(id)
);`

type Recurring struct {
	Interval string `json:"interval"`
	Until    uint   `json:"until"`
}

type Event struct {
	UUID           string    `json:"uuid"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	StartDate      uint      `json:"startDate"`
	EndDate        uint      `json:"endDate"`
	Recurring      Recurring `json:"recurring"`
	Type           string    `json:"type"`
	Participation  string    `json:"participation"`
	Attendance     uint      `json:"attendance"`
	RecurringEvent string
}

func (e *Event) Get() error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT name, startDate, endDate, type FROM %s WHERE uuid= ?", EVENTS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(e.UUID).Scan(&e.Name, &e.StartDate, &e.EndDate, &e.Type)
	return err
}

func (e *Event) GetAttendance() error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT COUNT(answer) FROM %s WHERE event_uuid= ? AND answer='yes'", PARTICIPATION_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(e.UUID).Scan(&e.Attendance)
	return err

}

func (e *Event) GetAll(start, count int) ([]Event, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT uuid, name, startDate, endDate, type FROM %s WHERE startDate > ? ORDER BY startDate LIMIT ?", EVENTS_TABLE), start, count)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	Events := []Event{}

	for rows.Next() {
		var e Event
		if err = rows.Scan(&e.UUID, &e.Name, &e.StartDate, &e.EndDate, &e.Type); err != nil {
			return nil, err
		}
		Events = append(Events, e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return Events, nil
}

func (e *Event) UpdateEvent() error {
	stmt, err := db.Prepare(fmt.Sprintf("Update %s SET name = ?, startDate = ?, endDate = ?, type = ? WHERE uuid= ?", EVENTS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(e.Name, e.StartDate, e.EndDate, e.Type, e.UUID)
	return err
}

func (e *Event) DeleteEvent() error {
	stmt, err := db.Prepare(fmt.Sprintf("DELETE FROM %s WHERE uuid= ?", EVENTS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(e.UUID)
	return err
}

func (e *Event) CreateEvent() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (uuid, name, startDate, endDate, recurringEvent, description, type) VALUES (?, ?, ?, ?, ?, ?, ?)", EVENTS_TABLE))
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(e.UUID, e.Name, e.StartDate, e.EndDate, e.RecurringEvent, e.Description, e.Type)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func (e *Event) GetUpcomingEventsWithoutNotification(eventType string) ([]Event, error) {
	rows, err := db.Query(fmt.Sprintf(
		"SELECT uuid, startDate FROM %s WHERE startDate > ? AND uuid NOT IN (SELECT objectUUID FROM notifications WHERE notificationType = '%s') ORDER BY startDate",
		EVENTS_TABLE, eventType), time.Now().Unix())
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	Events := []Event{}

	for rows.Next() {
		var e Event
		if err = rows.Scan(&e.UUID, &e.StartDate); err != nil {
			return nil, err
		}
		Events = append(Events, e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return Events, nil
}

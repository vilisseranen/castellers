package model

import (
	"fmt"
	"time"

	"github.com/vilisseranen/castellers/common"
)

// TODO: implement deleted flag on events

const EVENTS_TABLE = "events"

type Recurring struct {
	Interval string `json:"interval"`
	Until    uint   `json:"until"`
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
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
	Location       LatLng    `json:"location"`
	LocationName   string    `json:"locationName"`
	RecurringEvent string
}

func (e *Event) Get() error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT name, startDate, endDate, type, description, locationName, lat, lng FROM %s WHERE uuid= ?", EVENTS_TABLE))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	err = stmt.QueryRow(e.UUID).Scan(&e.Name, &e.StartDate, &e.EndDate, &e.Type, &e.Description, &e.LocationName, &e.Location.Lat, &e.Location.Lng)
	return err
}

func (e *Event) GetAttendance() error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT COUNT(answer) FROM %s WHERE event_uuid= ? AND answer='yes'", PARTICIPATION_TABLE))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	err = stmt.QueryRow(e.UUID).Scan(&e.Attendance)
	return err

}

func (e *Event) GetAll(start, count int) ([]Event, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT uuid, name, startDate, endDate, type FROM %s WHERE startDate > ? ORDER BY startDate LIMIT ?", EVENTS_TABLE), start, count)
	if err != nil {
		common.Fatal(err.Error())
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
	stmt, err := db.Prepare(fmt.Sprintf("Update %s SET name = ?, startDate = ?, endDate = ?, type = ?, description = ?, locationName = ?, lat = ?, lng = ? WHERE uuid= ?", EVENTS_TABLE))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		e.Name,
		e.StartDate,
		e.EndDate,
		e.Type,
		stringOrNull(e.Description),
		stringOrNull(e.LocationName),
		e.Location.Lat,
		e.Location.Lng,
		e.UUID)
	return err
}

func (e *Event) DeleteEvent() error {
	stmt, err := db.Prepare(fmt.Sprintf("DELETE FROM %s WHERE uuid= ?", EVENTS_TABLE))
	if err != nil {
		common.Fatal(err.Error())
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
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (uuid, name, startDate, endDate, recurringEvent, description, type, locationName, lat, lng) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", EVENTS_TABLE))
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		e.UUID,
		e.Name,
		e.StartDate,
		e.EndDate,
		e.RecurringEvent,
		stringOrNull(e.Description),
		e.Type,
		stringOrNull(e.LocationName),
		e.Location.Lat,
		e.Location.Lng)
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
		common.Fatal(err.Error())
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

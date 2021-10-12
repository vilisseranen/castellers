package model

import (
	"context"
	"database/sql"
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

func (e *Event) Get(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Event.Get")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("SELECT name, startDate, endDate, type, description, locationName, lat, lng FROM %s WHERE uuid= ? AND deleted=0", EVENTS_TABLE))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	var description, locationName sql.NullString // to manage possible NULL fields
	err = stmt.QueryRowContext(ctx, e.UUID).Scan(&e.Name, &e.StartDate, &e.EndDate, &e.Type, &description, &locationName, &e.Location.Lat, &e.Location.Lng)
	e.Description = nullToEmptyString(description)
	e.LocationName = nullToEmptyString(locationName)
	return err
}

func (e *Event) GetAttendance(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Event.GetAttendance")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("SELECT COUNT(answer) FROM %s WHERE event_uuid= ? AND answer='yes'", PARTICIPATION_TABLE))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	err = stmt.QueryRowContext(ctx, e.UUID).Scan(&e.Attendance)
	return err
}

func (e *Event) GetAll(ctx context.Context, page, limit int, pastEvents bool) ([]Event, error) {
	ctx, span := tracer.Start(ctx, "Event.GetAll")
	defer span.End()
	now := int(time.Now().Unix())
	offset := page * limit
	queryString := ""
	if pastEvents {
		queryString = fmt.Sprintf("SELECT uuid, name, startDate, endDate, type FROM %s WHERE endDate < ? AND deleted=0 ORDER BY startDate DESC LIMIT ? OFFSET ?", EVENTS_TABLE)
	} else {
		queryString = fmt.Sprintf("SELECT uuid, name, startDate, endDate, type FROM %s WHERE endDate >= ? AND deleted=0 ORDER BY startDate LIMIT ? OFFSET ?", EVENTS_TABLE)
	}
	rows, err := db.QueryContext(ctx, queryString, now, limit, offset)
	defer rows.Close()
	if err != nil {
		common.Fatal(err.Error())
	}

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

func (e *Event) UpdateEvent(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Event.UpdateEvent")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("Update %s SET name = ?, startDate = ?, endDate = ?, type = ?, description = ?, locationName = ?, lat = ?, lng = ? WHERE uuid= ?", EVENTS_TABLE))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	_, err = stmt.ExecContext(
		ctx,
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

func (e *Event) DeleteEvent(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Event.DeleteEvent")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("UPDATE %s SET deleted=1 WHERE uuid= ?", EVENTS_TABLE))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	_, err = stmt.ExecContext(ctx, e.UUID)
	return err
}

func (e *Event) CreateEvent(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Event.CreateEvent")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("INSERT INTO %s (uuid, name, startDate, endDate, recurringEvent, description, type, locationName, lat, lng) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", EVENTS_TABLE))
	defer stmt.Close()
	if err != nil {
		common.Error(err.Error())
		common.Error("%v\n", e)
		return err
	}
	_, err = stmt.ExecContext(
		ctx,
		e.UUID,
		e.Name,
		e.StartDate,
		e.EndDate,
		stringOrNull(e.RecurringEvent),
		stringOrNull(e.Description),
		e.Type,
		stringOrNull(e.LocationName),
		e.Location.Lat,
		e.Location.Lng)
	if err != nil {
		stmt.Close()
		common.Error(err.Error())
		common.Error("%v\n", e)
		return err
	}
	return err
}

func (e *Event) GetUpcomingEventsWithoutNotification(ctx context.Context, eventType string) ([]Event, error) {
	ctx, span := tracer.Start(ctx, "Event.GetUpcomingEventsWithoutNotification")
	defer span.End()
	rows, err := db.QueryContext(ctx, fmt.Sprintf(
		"SELECT uuid, startDate FROM %s WHERE startDate > ? AND uuid NOT IN (SELECT objectUUID FROM notifications WHERE notificationType = '%s') AND deleted=0 ORDER BY startDate",
		EVENTS_TABLE, eventType), time.Now().Unix())
	defer rows.Close()
	if err != nil {
		common.Fatal(err.Error())
	}

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

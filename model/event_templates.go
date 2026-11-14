package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vilisseranen/castellers/common"
)

const EVENT_TEMPLATES_TABLE = "event_templates"

var ErrEventTemplateNotFound = errors.New("event template not found")

// EventTemplate holds values proposed when an admin creates an event. Title
// is the name of the template itself; the other fields are those of an event.
// StartDate and EndDate are those of a sample event: only the time of day and
// the duration are meant to be used. A location of 0,0 means "none".
type EventTemplate struct {
	UUID            string `json:"uuid"`
	Title           string `json:"title"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	Description     string `json:"description"`
	LocationName    string `json:"locationName"`
	Location        LatLng `json:"location"`
	StartDate       uint   `json:"startDate"`
	EndDate         uint   `json:"endDate"`
	UniformRequired int    `json:"uniformRequired"`
}

const eventTemplateColumns = "uuid, title, name, type, description, location_name, lat, lng, start_date, end_date, uniform_required"

// GetEventTemplates returns all the templates, sorted by title.
func GetEventTemplates(ctx context.Context) ([]EventTemplate, error) {
	ctx, span := tracer.Start(ctx, "GetEventTemplates")
	defer span.End()

	rows, err := db.QueryContext(ctx, fmt.Sprintf(
		"SELECT %s FROM %s ORDER BY title COLLATE NOCASE", eventTemplateColumns, EVENT_TEMPLATES_TABLE))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := []EventTemplate{}
	for rows.Next() {
		var t EventTemplate
		if err := rows.Scan(&t.UUID, &t.Title, &t.Name, &t.Type, &t.Description, &t.LocationName,
			&t.Location.Lat, &t.Location.Lng, &t.StartDate, &t.EndDate, &t.UniformRequired); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// EventTemplateTitleTaken tells whether another template already has this
// title, whatever the case. exceptUUID is the template being saved, if any.
func EventTemplateTitleTaken(ctx context.Context, title, exceptUUID string) (bool, error) {
	ctx, span := tracer.Start(ctx, "EventTemplateTitleTaken")
	defer span.End()

	var count int
	err := db.QueryRowContext(ctx, fmt.Sprintf(
		"SELECT COUNT(uuid) FROM %s WHERE title = ? COLLATE NOCASE AND uuid != ?", EVENT_TEMPLATES_TABLE),
		title, exceptUUID).Scan(&count)
	return count > 0, err
}

// Create saves a new template and sets its UUID.
func (t *EventTemplate) Create(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "EventTemplate.Create")
	defer span.End()

	t.UUID = common.GenerateUUID()
	_, err := db.ExecContext(ctx, fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", EVENT_TEMPLATES_TABLE, eventTemplateColumns),
		t.UUID, t.Title, t.Name, t.Type, t.Description, t.LocationName,
		t.Location.Lat, t.Location.Lng, t.StartDate, t.EndDate, t.UniformRequired)
	return err
}

// Update replaces the template that has this UUID, or returns
// ErrEventTemplateNotFound.
func (t *EventTemplate) Update(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "EventTemplate.Update")
	defer span.End()

	result, err := db.ExecContext(ctx, fmt.Sprintf(
		"UPDATE %s SET title = ?, name = ?, type = ?, description = ?, location_name = ?, lat = ?, lng = ?, "+
			"start_date = ?, end_date = ?, uniform_required = ? WHERE uuid = ?", EVENT_TEMPLATES_TABLE),
		t.Title, t.Name, t.Type, t.Description, t.LocationName,
		t.Location.Lat, t.Location.Lng, t.StartDate, t.EndDate, t.UniformRequired, t.UUID)
	if err != nil {
		return err
	}
	return eventTemplateAffected(result)
}

// DeleteEventTemplate removes a template, or returns
// ErrEventTemplateNotFound. Events created from it are not touched: they keep
// no link to the template.
func DeleteEventTemplate(ctx context.Context, uuid string) error {
	ctx, span := tracer.Start(ctx, "DeleteEventTemplate")
	defer span.End()

	result, err := db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE uuid = ?", EVENT_TEMPLATES_TABLE), uuid)
	if err != nil {
		return err
	}
	return eventTemplateAffected(result)
}

func eventTemplateAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrEventTemplateNotFound
	}
	return nil
}

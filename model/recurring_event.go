package model

import (
	"context"
	"fmt"

	"github.com/vilisseranen/castellers/common"
	"go.elastic.co/apm"
)

const RECURRING_EVENTS_TABLE = "recurring_events"

type RecurringEvent struct {
	UUID        string
	Name        string
	Description string
	Interval    string
}

func (r *RecurringEvent) Get(ctx context.Context) error {
	span, ctx := apm.StartSpan(ctx, "RecurringEvent.Get", APM_SPAN_TYPE_REQUEST)
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("SELECT name, description, interval FROM %s WHERE uuid= ?", RECURRING_EVENTS_TABLE))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	err = stmt.QueryRowContext(ctx, r.UUID).Scan(&r.Name, &r.Description, &r.Interval)
	return err
}

func (r *RecurringEvent) CreateRecurringEvent(ctx context.Context) error {
	span, ctx := apm.StartSpan(ctx, "RecurringEvent.CreateRecurringEvent", APM_SPAN_TYPE_REQUEST)
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("INSERT INTO %s (uuid, name, description, interval) VALUES (?, ?, ?, ?)", RECURRING_EVENTS_TABLE))
	defer stmt.Close()
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, r.UUID, r.Name, r.Description, r.Interval)
	if err != nil {
		stmt.Close()
		common.Error(err.Error())
		common.Error("%v\n", r)
		return err
	}
	return err
}

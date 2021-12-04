package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/vilisseranen/castellers/common"
)

const PARTICIPATION_TABLE = "participation"

type Participation struct {
	EventUUID  string `json:"eventUuid"`
	MemberUUID string `json:"memberUuid"`
	Answer     string `json:"answer"`
	Presence   string `json:"presence"`
}

// A member will say if he or she participates BEFORE the event:
// We always insert in the table
func (p *Participation) Participate(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Participation.Participate")
	defer span.End()

	// Check if a participation already exists
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("SELECT count(*) FROM %s WHERE member_uuid= ? AND event_uuid= ?", PARTICIPATION_TABLE))
	defer stmt.Close()
	if err != nil {
		return err
	}
	c := 0
	err = stmt.QueryRowContext(ctx, p.MemberUUID, p.EventUUID).Scan(&c)
	if err != nil {
		return err
	}
	if c == 0 {
		stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO %s (event_uuid, member_uuid, answer, presence) VALUES (?, ?, ?, ?)", PARTICIPATION_TABLE))
		defer stmt.Close()
		if err != nil {
			return err
		}
		_, err = stmt.Exec(p.EventUUID, p.MemberUUID, stringOrNull(p.Answer), "")
		if err != nil {
			return err
		}
	} else if c == 1 {
		stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s SET answer = ? WHERE event_uuid= ? AND member_uuid= ?", PARTICIPATION_TABLE))
		defer stmt.Close()
		if err != nil {
			return err
		}
		_, err = stmt.Exec(stringOrNull(p.Answer), p.EventUUID, p.MemberUUID)
		if err != nil {
			return err
		}
	}
	return err
}

func (p *Participation) GetParticipation(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Participation.GetParticipation")
	defer span.End()
	// Check if a participation already exists
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("SELECT answer, presence FROM %s WHERE member_uuid= ? AND event_uuid= ?", PARTICIPATION_TABLE))
	defer stmt.Close()
	if err != nil {
		return err
	}
	var answer, presence sql.NullString // to manage possible NULL fields
	err = stmt.QueryRowContext(ctx, p.MemberUUID, p.EventUUID).Scan(&answer, &presence)
	p.Answer = nullToEmptyString(answer)
	p.Presence = nullToEmptyString(presence)
	return err
}

// A member might be present even if he is not registered for the event
func (p *Participation) Present(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "GetEvent")
	defer span.End()
	// Check if a participation already exists
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("SELECT count(*) FROM %s WHERE member_uuid= ? AND event_uuid= ?", PARTICIPATION_TABLE))
	defer stmt.Close()
	if err != nil {
		return err
	}
	c := 0
	err = stmt.QueryRowContext(ctx, p.MemberUUID, p.EventUUID).Scan(&c)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	// c should be 0 (the member did not register) or 1 (the member did register)
	if c == 0 {
		stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (event_uuid, member_uuid, presence, answer) VALUES (?, ?, ?, ?)", PARTICIPATION_TABLE))
		defer stmt.Close()
		if err != nil {
			tx.Rollback()
			common.Error("%v", err)
			return err
		}
		_, err = stmt.Exec(p.EventUUID, p.MemberUUID, stringOrNull(p.Presence), "")
		if err != nil {
			tx.Rollback()
			common.Error("%v\n", err)
			return err
		}
	} else if c == 1 {
		stmt, err := tx.Prepare(fmt.Sprintf("UPDATE %s SET presence= ? WHERE event_uuid= ? AND member_uuid= ?", PARTICIPATION_TABLE))
		defer stmt.Close()
		if err != nil {
			tx.Rollback()
			common.Error("%v\n", err)
			return err
		}
		_, err = stmt.Exec(stringOrNull(p.Presence), p.EventUUID, p.MemberUUID)
		if err != nil {
			tx.Rollback()
			common.Error("%v\n", err)
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		common.Error("%v\n", err)
		tx.Rollback()
	}
	return err
}

func (m *Member) GetMemberLastParticipation(ctx context.Context) (Event, error) {
	ctx, span := tracer.Start(ctx, "Participation.GetMemberLastParticipation")
	defer span.End()

	var e Event

	query := fmt.Sprintf(
		`SELECT e.uuid FROM %s AS e LEFT JOIN %s AS p ON e.uuid = p.event_uuid
		 WHERE p.member_uuid = ? AND (p.answer = '%s' OR p.presence = '%s') AND e.deleted = 0
		 ORDER BY e.startDate DESC LIMIT 1;`, EVENTS_TABLE, PARTICIPATION_TABLE, common.AnswerYes, common.AnswerYes)
	common.Debug("SQL query: " + query)
	stmt, err := db.PrepareContext(ctx, query)
	defer stmt.Close()
	if err != nil {
		return e, err
	}
	err = stmt.QueryRowContext(ctx, m.UUID).Scan(&e.UUID)
	if err != nil {
		return e, err
	}
	err = e.Get(ctx)
	if err != nil {
		return e, err
	}
	return e, nil
}

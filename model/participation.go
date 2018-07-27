package model

import (
	"fmt"
)

const PARTICIPATION_TABLE = "participation"

const ParticipationTableCreationQuery = `CREATE TABLE IF NOT EXISTS participation
(
	member_uuid INTEGER NOT NULL,
	event_uuid INTEGER NOT NULL,
    answer TEXT,
	presence TEXT,
	PRIMARY KEY (member_uuid, event_uuid)
);`

type Participation struct {
	EventUUID  string `json:"eventUuid"`
	MemberUUID string `json:"memberUuid"`
	Answer     string `json:"answer"`
	Presence   string `json:"presence"`
}

// A member will say if he or she participes BEFORE the event:
// We always insert in the table
func (p *Participation) Participate() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (event_uuid, member_uuid, answer) VALUES (?, ?, ?)", PARTICIPATION_TABLE))
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(p.EventUUID, p.MemberUUID, p.Answer)
	if err != nil {
		return err
	}
	tx.Commit()
	return err
}

// A member might be present even if he is not registered for the event:
// We can check presence only when the event STARTS
func (p *Participation) Present() error {
	// Check if a participation already exists
	stmt, err := db.Prepare(fmt.Sprintf("SELECT count(*) FROM %s WHERE member_uuid= ? AND event_uuid= ?", PARTICIPATION_TABLE))
	defer stmt.Close()
	c := 0
	err = stmt.QueryRow(p.MemberUUID, p.EventUUID).Scan(&c)
	if err != nil {
		return err
	}

	// c should be 0 (the member did not register) or 1 (the member did register)
	if c == 0 {
		stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO %s (event_uuid, member_uuid, presence) VALUES (?, ?, ?)", PARTICIPATION_TABLE))
		if err != nil {
			return err
		}
		_, err = stmt.Exec(p.EventUUID, p.MemberUUID, p.Presence)
		if err != nil {
			return err
		}
	} else if c == 1 {
		stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s SET presence= ? WHERE event_uuid= ? AND member_uuid= ?", PARTICIPATION_TABLE))
		if err != nil {
			return err
		}
		_, err = stmt.Exec(p.Presence, p.EventUUID, p.MemberUUID)
		if err != nil {
			return err
		}
	}
	return err
}

package model

import (
	"fmt"
)

const PARTICIPATION_TABLE = "participation"

const ParticipationTableCreationQuery = `CREATE TABLE IF NOT EXISTS participation
(
	member_uuid INTEGER NOT NULL,
	event_uuid INTEGER NOT NULL,
  answer TEXT NOT NULL,
	PRIMARY KEY (member_uuid, event_uuid)
);`

type Participation struct {
	EventUUID  string `json:"eventUuid"`
	MemberUUID string `json:"memberUuid"`
	Answer     string `json:"answer"`
}

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

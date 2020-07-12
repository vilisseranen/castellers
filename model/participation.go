package model

import (
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
func (p *Participation) Participate() error {
	// Check if a participation already exists
	stmt, err := db.Prepare(fmt.Sprintf("SELECT count(*) FROM %s WHERE member_uuid= ? AND event_uuid= ?", PARTICIPATION_TABLE))
	defer stmt.Close()
	if err != nil {
		return err
	}
	c := 0
	err = stmt.QueryRow(p.MemberUUID, p.EventUUID).Scan(&c)
	if err != nil {
		return err
	}
	if c == 0 {
		stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO %s (event_uuid, member_uuid, answer, presence) VALUES (?, ?, ?, ?)", PARTICIPATION_TABLE))
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(p.EventUUID, p.MemberUUID, stringOrNull(p.Answer), "")
		if err != nil {
			return err
		}
	} else if c == 1 {
		stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s SET answer = ? WHERE event_uuid= ? AND member_uuid= ?", PARTICIPATION_TABLE))
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(stringOrNull(p.Answer), p.EventUUID, p.MemberUUID)
		if err != nil {
			return err
		}
	}
	return err
}

func (p *Participation) GetParticipation() error {
	// Check if a participation already exists
	stmt, err := db.Prepare(fmt.Sprintf("SELECT answer, presence FROM %s WHERE member_uuid= ? AND event_uuid= ?", PARTICIPATION_TABLE))
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(p.MemberUUID, p.EventUUID).Scan(&p.Answer, &p.Presence)
	return err
}

// A member might be present even if he is not registered for the event
func (p *Participation) Present() error {
	// Check if a participation already exists
	stmt, err := db.Prepare(fmt.Sprintf("SELECT count(*) FROM %s WHERE member_uuid= ? AND event_uuid= ?", PARTICIPATION_TABLE))
	defer stmt.Close()
	if err != nil {
		return err
	}
	c := 0
	err = stmt.QueryRow(p.MemberUUID, p.EventUUID).Scan(&c)
	if err != nil {
		return err
	}

	// c should be 0 (the member did not register) or 1 (the member did register)
	if c == 0 {
		stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO %s (event_uuid, member_uuid, presence, answer) VALUES (?, ?, ?, ?)", PARTICIPATION_TABLE))
		if err != nil {
			common.Error("%v", err)
			return err
		}
		_, err = stmt.Exec(p.EventUUID, p.MemberUUID, stringOrNull(p.Presence), "")
		if err != nil {
			return err
		}
	} else if c == 1 {
		stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s SET presence= ? WHERE event_uuid= ? AND member_uuid= ?", PARTICIPATION_TABLE))
		if err != nil {
			return err
		}
		_, err = stmt.Exec(stringOrNull(p.Presence), p.EventUUID, p.MemberUUID)
		if err != nil {
			return err
		}
	}
	return err
}

package model

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/vilisseranen/castellers/common"
)

// Tables names
const EventS_TABLE = "events"
const MEMBERS_TABLE = "members"
const ADMINS_TABLE = "admins"
const PARTICIPATION_TABLE = "participation"

// Tables creation queries
const EventsTableCreationQuery = `CREATE TABLE IF NOT EXISTS events
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	startDate TEXT NOT NULL,
	endDate TEXT NOT NULL,
	description TEXT,
	uuid TEXT NOT NULL,
	CONSTRAINT uuid_unique UNIQUE (uuid)
);`
const AdminsTableCreationQuery = `CREATE TABLE IF NOT EXISTS admins
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uuid TEXT NOT NULL,
	CONSTRAINT uuid_unique UNIQUE (uuid)
);`
const MembersTableCreationQuery = `CREATE TABLE IF NOT EXISTS members
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uuid TEXT NOT NULL,
	name TEXT NOT NULL,
	extra TEXT,
	CONSTRAINT uuid_unique UNIQUE (uuid)
);`

const ParticipationTableCreationQuery = `CREATE TABLE IF NOT EXISTS participation
(
	member_uuid INTEGER NOT NULL,
	event_uuid INTEGER NOT NULL,
  answer TEXT NOT NULL,
	PRIMARY KEY (member_uuid, event_uuid)
);`

type Event struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type Admin struct {
	UUID string `json:"uuid"`
}

type Member struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Extra string `json:"extra"`
}

type Participation struct {
	EventUUID  string `json:"eventUuid"`
	MemberUUID string `json:"memberUuid"`
	Answer     string `json:"answer"`
}

type Entity interface {
	Get(db *sql.DB) error
}

func (a *Admin) Get(db *sql.DB) error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT uuid FROM %s WHERE uuid= ?", ADMINS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(a.UUID).Scan(&a.UUID)
	return err
}

func (e *Event) Get(db *sql.DB) error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT name, startDate, endDate FROM %s WHERE uuid= ?", EventS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(e.UUID).Scan(&e.Name, &e.StartDate, &e.EndDate)
	return err
}

// TODO: this needs to be implemented by Get
func (e *Event) GetEvents(db *sql.DB, start, count int) ([]Event, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT uuid, name, startDate, endDate FROM %s LIMIT ? OFFSET ?", EventS_TABLE), count, start)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	Events := []Event{}

	for rows.Next() {
		var e Event
		if err = rows.Scan(&e.UUID, &e.Name, &e.StartDate, &e.EndDate); err != nil {
			return nil, err
		}
		Events = append(Events, e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return Events, nil
}

func (e *Event) UpdateEvent(db *sql.DB) error {
	stmt, err := db.Prepare(fmt.Sprintf("Update %s SET name = ?, startDate = ?, endDate = ? WHERE uuid= ?", EventS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(e.Name, e.StartDate, e.EndDate, e.UUID)
	return err
}

func (e *Event) DeleteEvent(db *sql.DB) error {
	stmt, err := db.Prepare(fmt.Sprintf("DELETE FROM %s WHERE uuid= ?", EventS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(e.UUID)
	return err
}

func (e *Event) CreateEvent(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (uuid, name, startDate, endDate) VALUES (?, ?, ?, ?)", EventS_TABLE))
	if err != nil {
		return err
	}
	defer stmt.Close()
	e.UUID = common.GenerateUUID()
	_, err = stmt.Exec(e.UUID, e.Name, e.StartDate, e.EndDate)
	if err != nil {
		return err
	}
	tx.Commit()
	return err
}

func (m *Member) CreateMember(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (uuid, name, extra) VALUES (?, ?, ?)", MEMBERS_TABLE))
	if err != nil {
		return err
	}
	defer stmt.Close()
	m.UUID = common.GenerateUUID()
	_, err = stmt.Exec(m.UUID, m.Name, m.Extra)
	if err != nil {
		return err
	}
	tx.Commit()
	return err
}

func (m *Member) Get(db *sql.DB) error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT name, extra FROM %s WHERE uuid= ?", MEMBERS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(m.UUID).Scan(&m.Name, &m.Extra)
	return err
}

func (p *Participation) Participate(db *sql.DB) error {
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

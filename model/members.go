package model

import (
	"fmt"
	"log"
)

const MEMBERS_TABLE = "members"

const MEMBER_TYPE_ADMIN = "admin"
const MEMBER_TYPE_MEMBER = "member"

const MembersTableCreationQuery = `CREATE TABLE IF NOT EXISTS members
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uuid TEXT NOT NULL,
	firstName TEXT NOT NULL,
	lastName TEXT NOT NULL,
	extra TEXT,
	roles TEXT,
	type TEXT NOT NULL,
	email TEXT NOT NULL,
	code TEXT NOT NULL,
	activated INTEGER NOT NULL DEFAULT 0,
	CONSTRAINT uuid_unique UNIQUE (uuid)
);`

type Member struct {
	UUID      string `json:"uuid"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Roles     string `json:"roles"`
	Extra     string `json:"extra"`
	Type      string `json:"type"`
	Email     string `json:"email"`
	Code      string `json:"-"`
	Activated int    `json:"activated"`
}

func (m *Member) CreateMember() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf(
		"INSERT INTO %s (uuid, firstName, lastName, roles, extra, type, email, code) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		MEMBERS_TABLE))
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		stringOrNull(m.UUID),
		stringOrNull(m.FirstName),
		stringOrNull(m.LastName),
		stringOrNull(m.Roles),
		stringOrNull(m.Extra),
		stringOrNull(m.Type),
		stringOrNull(m.Email),
		stringOrNull(m.Code))
	if err != nil {
		fmt.Printf("%v\n", m)
		return err
	}
	tx.Commit()
	return err
}

func (m *Member) Get() error {
	stmt, err := db.Prepare(fmt.Sprintf(
		"SELECT firstName, lastName, roles, extra, type, email, code FROM %s WHERE uuid= ?",
		MEMBERS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(m.UUID).Scan(&m.FirstName, &m.LastName, &m.Roles, &m.Extra, &m.Type, &m.Email, &m.Code)
	return err
}

func (m *Member) GetAll() ([]Member, error) {
	rows, err := db.Query(fmt.Sprintf(
		"SELECT uuid, firstName, lastName, roles, extra, type, email, code FROM %s",
		MEMBERS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	members := []Member{}

	for rows.Next() {
		var m Member
		if err = rows.Scan(&m.UUID, &m.FirstName, &m.LastName, &m.Roles, &m.Extra, &m.Type, &m.Email, &m.Code); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

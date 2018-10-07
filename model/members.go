package model

import (
	"fmt"
	"log"
	"strings"
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
	extra TEXT NOT NULL,
	roles TEXT NOT NULL,
	type TEXT NOT NULL,
	email TEXT NOT NULL,
	code TEXT NOT NULL,
	activated INTEGER NOT NULL DEFAULT 0,
	deleted INTEGER NOT NULL DEFAULT 0,
	language TEXT NOT NULL DEFAULT 'fr',
	CONSTRAINT uuid_unique UNIQUE (uuid)
);`

type Member struct {
	UUID      string   `json:"uuid"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Roles     []string `json:"roles"`
	Extra     string   `json:"extra"`
	Type      string   `json:"type"`
	Email     string   `json:"email"`
	Code      string   `json:"-"`
	Activated int      `json:"activated"`
	Deleted   int      `json:"-"`
	Language  string   `json:"language"`
}

func (m *Member) CreateMember() error {
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("%v\n", m)
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf(
		"INSERT INTO %s (uuid, firstName, lastName, roles, extra, type, email, code, language) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		MEMBERS_TABLE))
	if err != nil {
		fmt.Printf("%v\n", m)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		stringOrNull(m.UUID),
		stringOrNull(m.FirstName),
		stringOrNull(m.LastName),
		stringOrNull(strings.Join(m.Roles, ",")),
		stringOrNull(m.Extra),
		stringOrNull(m.Type),
		stringOrNull(m.Email),
		stringOrNull(m.Code),
		stringOrNull(m.Language))
	if err != nil {
		fmt.Printf("%v\n", m)
		return err
	}
	tx.Commit()
	return err
}

func (m *Member) EditMember() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf(
		"UPDATE %s SET firstName=?, lastName=?, roles=?, extra=?, type=?, email=?, language=? WHERE uuid=?",
		MEMBERS_TABLE))
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		stringOrNull(m.FirstName),
		stringOrNull(m.LastName),
		stringOrNull(strings.Join(m.Roles, ",")),
		stringOrNull(m.Extra),
		stringOrNull(m.Type),
		stringOrNull(m.Email),
		stringOrNull(m.Language),
		stringOrNull(m.UUID))
	if err != nil {
		fmt.Printf("%v\n", m)
		return err
	}
	tx.Commit()
	return err
}

func (m *Member) Get() error {
	stmt, err := db.Prepare(fmt.Sprintf(
		"SELECT firstName, lastName, roles, extra, type, email, code, activated, language FROM %s WHERE uuid= ? AND deleted=0",
		MEMBERS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var rolesAsString string
	err = stmt.QueryRow(m.UUID).Scan(&m.FirstName, &m.LastName, &rolesAsString, &m.Extra, &m.Type, &m.Email, &m.Code, &m.Activated, &m.Language)
	m.Roles = strings.Split(rolesAsString, ",")
	m.sanitizeEmptyRoles()
	return err
}

func (m *Member) GetAll() ([]Member, error) {
	rows, err := db.Query(fmt.Sprintf(
		"SELECT uuid, firstName, lastName, roles, extra, type, email, code, activated, language FROM %s WHERE deleted=0",
		MEMBERS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	members := []Member{}

	for rows.Next() {
		var m Member
		var rolesAsString string
		if err = rows.Scan(&m.UUID, &m.FirstName, &m.LastName, &rolesAsString, &m.Extra, &m.Type, &m.Email, &m.Code, &m.Activated, &m.Language); err != nil {
			return nil, err
		}
		m.Roles = strings.Split(rolesAsString, ",")
		m.sanitizeEmptyRoles()
		members = append(members, m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

func (m *Member) DeleteMember() error {
	stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s SET deleted=1 WHERE uuid=?",
		MEMBERS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.UUID)
	return err
}

func (m *Member) sanitizeEmptyRoles() {
	if len(m.Roles) == 1 && m.Roles[0] == "" {
		m.Roles = []string{}
	}
	return
}

func (m *Member) Activate() error {
	stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s SET activated = 1 WHERE uuid= ?", MEMBERS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.UUID)
	return err
}

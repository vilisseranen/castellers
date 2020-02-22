package model

import (
	"fmt"
	"log"
	"strings"
)

const MembersTable = "members"
const MemberTypeAdmin = "admin"

const MemberTypeMember = "member"

const MembersTableCreationQuery = `CREATE TABLE IF NOT EXISTS members
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uuid TEXT NOT NULL,
	firstName TEXT NOT NULL,
	lastName TEXT NOT NULL,
	height TEXT NOT NULL,
	weight TEXT NOT NULL,
	extra TEXT NOT NULL,
	roles TEXT NOT NULL,
	type TEXT NOT NULL,
	email TEXT NOT NULL,
	contact TEXT NOT NULL,
	code TEXT NOT NULL,
	activated INTEGER NOT NULL DEFAULT 0,
	subscribed INTEGER NOT NULL DEFAULT 0,
	deleted INTEGER NOT NULL DEFAULT 0,
	language TEXT NOT NULL DEFAULT 'fr',
	CONSTRAINT uuid_unique UNIQUE (uuid)
);`

type Member struct {
	UUID          string   `json:"uuid"`
	FirstName     string   `json:"firstName"`
	LastName      string   `json:"lastName"`
	Height        string   `json:"height"`
	Weight        string   `json:"weight"`
	Roles         []string `json:"roles"`
	Extra         string   `json:"extra"`
	Type          string   `json:"type"`
	Email         string   `json:"email"`
	Contact       string   `json:"contact"`
	Code          string   `json:"-"`
	Activated     int      `json:"activated"`
	Subscribed    int      `json:"subscribed"`
	Deleted       int      `json:"-"`
	Language      string   `json:"language"`
	Participation string   `json:"participation"`
}

func (m *Member) CreateMember() error {
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("%v\n", m)
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf(
		"INSERT INTO %s (uuid, firstName, lastName, height, weight, roles, extra, type, email, contact, code, language) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		MembersTable))
	if err != nil {
		fmt.Printf("%v\n", m)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		stringOrNull(m.UUID),
		stringOrNull(m.FirstName),
		stringOrNull(m.LastName),
		stringOrNull(m.Height),
		stringOrNull(m.Weight),
		stringOrNull(strings.Join(m.Roles, ",")),
		stringOrNull(m.Extra),
		stringOrNull(m.Type),
		stringOrNull(m.Email),
		stringOrNull(m.Contact),
		stringOrNull(m.Code),
		stringOrNull(m.Language))
	if err != nil {
		fmt.Printf("%v\n", m)
		return err
	}
	err = tx.Commit()
	return err
}

func (m *Member) EditMember() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf(
		"UPDATE %s SET firstName=?, lastName=?, height=?, weight=?, roles=?, extra=?, type=?, email=?, contact=?, language=?, subscribed=? WHERE uuid=?",
		MembersTable))
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		stringOrNull(m.FirstName),
		stringOrNull(m.LastName),
		stringOrNull(m.Height),
		stringOrNull(m.Weight),
		stringOrNull(strings.Join(m.Roles, ",")),
		stringOrNull(m.Extra),
		stringOrNull(m.Type),
		stringOrNull(m.Email),
		stringOrNull(m.Contact),
		stringOrNull(m.Language),
		m.Subscribed,
		stringOrNull(m.UUID))
	if err != nil {
		fmt.Printf("%v\n", m)
		return err
	}
	err = tx.Commit()
	return err
}

func (m *Member) Get() error {
	stmt, err := db.Prepare(fmt.Sprintf(
		"SELECT firstName, lastName, height, weight, roles, extra, type, email, contact, code, activated, subscribed, language FROM %s WHERE uuid= ? AND deleted=0",
		MembersTable))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var rolesAsString string
	err = stmt.QueryRow(m.UUID).Scan(&m.FirstName, &m.LastName, &m.Height, &m.Weight, &rolesAsString, &m.Extra, &m.Type, &m.Email, &m.Contact, &m.Code, &m.Activated, &m.Subscribed, &m.Language)
	m.Roles = strings.Split(rolesAsString, ",")
	m.sanitizeEmptyRoles()
	return err
}

func (m *Member) GetAll() ([]Member, error) {
	rows, err := db.Query(fmt.Sprintf(
		"SELECT uuid, firstName, lastName, height, weight, roles, extra, type, email, contact, code, activated, subscribed, language FROM %s WHERE deleted=0",
		MembersTable))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	members := []Member{}

	for rows.Next() {
		var m Member
		var rolesAsString string
		if err = rows.Scan(&m.UUID, &m.FirstName, &m.LastName, &m.Height, &m.Height, &rolesAsString, &m.Extra, &m.Type, &m.Email, &m.Contact, &m.Code, &m.Activated, &m.Subscribed, &m.Language); err != nil {
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
		MembersTable))
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
	stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s SET activated = 1 WHERE uuid= ?", MembersTable))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.UUID)
	return err
}

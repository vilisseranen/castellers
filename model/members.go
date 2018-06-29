package model

import (
	"fmt"
	"log"

	"github.com/vilisseranen/castellers/common"
)

const MEMBERS_TABLE = "members"

const MEMBER_TYPE_ADMIN = "admin"
const MEMBER_TYPE_MEMBER = "member"

const MembersTableCreationQuery = `CREATE TABLE IF NOT EXISTS members
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uuid TEXT NOT NULL,
	name TEXT NOT NULL,
	extra TEXT,
	type TEXT NOT NULL,
	code TEXT NOT NULL,
	CONSTRAINT uuid_unique UNIQUE (uuid)
);`

type Member struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Extra string `json:"extra"`
	Type  string `json:"type"`
	Code  string `json:"-"`
}

func (m *Member) CreateMember() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	m.Code = common.GenerateCode()
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (uuid, name, extra, type, code) VALUES (?, ?, ?, ?, ?)", MEMBERS_TABLE))
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.UUID, m.Name, m.Extra, m.Type, m.Code)
	if err != nil {
		return err
	}
	tx.Commit()
	return err
}

func (m *Member) Get() error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT name, extra, type, code FROM %s WHERE uuid= ?", MEMBERS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(m.UUID).Scan(&m.Name, &m.Extra, &m.Type, &m.Code)
	return err
}

func (m *Member) GetAll(start, count int) ([]Member, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT uuid, name, extra, type FROM %s LIMIT ? OFFSET ?", MEMBERS_TABLE), count, start)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	members := []Member{}

	for rows.Next() {
		var member Member
		if err = rows.Scan(&m.UUID, &m.Name, &m.Extra, &m.Type); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

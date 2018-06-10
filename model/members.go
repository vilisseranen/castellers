package model

import (
	"fmt"
	"log"

	"github.com/vilisseranen/castellers/common"
)

const MEMBERS_TABLE = "members"

const MembersTableCreationQuery = `CREATE TABLE IF NOT EXISTS members
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uuid TEXT NOT NULL,
	name TEXT NOT NULL,
	extra TEXT,
	CONSTRAINT uuid_unique UNIQUE (uuid)
);`

type Member struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Extra string `json:"extra"`
}

func (m *Member) CreateMember() error {
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

func (m *Member) Get() error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT name, extra FROM %s WHERE uuid= ?", MEMBERS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(m.UUID).Scan(&m.Name, &m.Extra)
	return err
}

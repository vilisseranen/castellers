package model

import (
	"fmt"
	"log"
)

const ADMINS_TABLE = "admins"

const AdminsTableCreationQuery = `CREATE TABLE IF NOT EXISTS admins
(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uuid TEXT NOT NULL,
	CONSTRAINT uuid_unique UNIQUE (uuid)
);`

type Admin struct {
	UUID string `json:"uuid"`
}

func (a *Admin) Get() error {
	stmt, err := db.Prepare(fmt.Sprintf("SELECT uuid FROM %s WHERE uuid= ?", ADMINS_TABLE))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(a.UUID).Scan(&a.UUID)
	return err
}

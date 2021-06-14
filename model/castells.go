package model

import (
	"fmt"

	"github.com/vilisseranen/castellers/common"
)

const (
	CASTELLTYPESVIEW   = "castell_types_view"
	CASTELLTYPESTABLE  = "castell_types"
	CASTELLMODELSTABLE = "castell_models"
)

type CastellType struct {
	Name      string            `json:"name"`
	Positions []CastellPosition `json:"positions"`
}

type CastellPosition struct {
	Name   string `json:"name"`
	Column string `json:"column"`
	Cordon string `json:"cordon"`
	Part   string `json:"part"`
}

type CastellModel struct {
	Name            string                   `json:"name"`
	Type            string                   `json:"type"` // Will be the name of the castell type, ie: 3d6
	PositionMembers []CastellPositionMembers `json:"position_members"`
}

type CastellPositionMembers struct {
	MemberUUID string `json:"member_uuid"`
	PositionID int    `json:"position_id"`
}

func (c *CastellType) Get() error {
	rows, err := db.Query(fmt.Sprintf(
		"SELECT position_name, position_column, position_cordon, position_part FROM %s WHERE castell_name= ?",
		CASTELLTYPESVIEW), c.Name)
	if err != nil {
		common.Fatal(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var p CastellPosition
		if err = rows.Scan(&p.Name, &p.Column, &p.Cordon, &p.Part); err != nil {
			return err
		}
		c.Positions = append(c.Positions, p)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return err
}

func (c *CastellType) GetTypeList() ([]string, error) {
	castell_types := []string{}
	rows, err := db.Query(fmt.Sprintf("SELECT name FROM %s", CASTELLTYPESTABLE))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, err
		}
		castell_types = append(castell_types, name)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return castell_types, err
}

func (c *CastellModel) Get() error {
	return nil
}

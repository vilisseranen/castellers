package model

import (
	"fmt"

	"github.com/vilisseranen/castellers/common"
)

const (
	CASTELLTYPESVIEW            = "castell_types_view"
	CASTELLTYPESTABLE           = "castell_types"
	CASTELLPOSITIONSTABLE       = "castell_positions"
	CASTELLMODELSTABLE          = "castell_models"
	CASTELLMODELSVIEW           = "castell_models_view"
	CASTELLMEMBERPOSITIONSTABLE = "castell_members_positions"
)

type CastellType struct {
	Name      string            `json:"name"`
	Positions []CastellPosition `json:"positions"`
}

type CastellPosition struct {
	Name   string `json:"name"`
	Column int    `json:"column"`
	Cordon int    `json:"cordon"`
	Part   string `json:"part"`
}

type CastellModel struct {
	UUID            string                   `json:"uuid"`
	Name            string                   `json:"name"`
	Type            string                   `json:"type"` // Will be the name of the castell type, ie: 3d6
	PositionMembers []CastellPositionMembers `json:"position_members,omitempty"`
}

type CastellPositionMembers struct {
	MemberUUID string          `json:"member_uuid"`
	Position   CastellPosition `json:"position"`
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

func (c *CastellModel) Create() error {
	tx, err := db.Begin()
	if err != nil {
		common.Error("%v\n", err)
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf(
		"INSERT INTO %s (uuid, name, castell_type_name) VALUES (?, ?, ?)",
		CASTELLMODELSTABLE))
	defer stmt.Close()
	if err != nil {
		tx.Rollback()
		common.Error("%v\n", err)
		return err
	}
	_, err = stmt.Exec(
		c.UUID,
		c.Name,
		c.Type,
	)
	if err != nil {
		tx.Rollback()
		common.Error("%v\n", err)
		return err
	}
	// For each position, add it in the table
	for _, member := range c.PositionMembers {
		stmt, err = tx.Prepare(fmt.Sprintf(
			"INSERT INTO %s "+
				"(castell_model_id, castell_position_id, member_id) VALUES ("+
				"(SELECT id FROM %s WHERE uuid = ?), "+
				"(SELECT id FROM %s WHERE name = ? AND column = ? AND cordon = ? AND part = ?), "+
				"(SELECT id FROM %s WHERE uuid = ?))",
			CASTELLMEMBERPOSITIONSTABLE, CASTELLMODELSTABLE, CASTELLPOSITIONSTABLE, MEMBERSTABLE))
		defer stmt.Close()
		if err != nil {
			tx.Rollback()
			common.Error("%v\n", err)
			return err
		}
		_, err = stmt.Exec(
			c.UUID,
			member.Position.Name,
			member.Position.Column,
			member.Position.Cordon,
			member.Position.Part,
			member.MemberUUID,
		)
		if err != nil {
			tx.Rollback()
			common.Error("%v\n", err)
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		err = tx.Rollback()
		common.Error("%v\n", err)
	}
	return err
}

func (c *CastellModel) Edit() error {
	// Delete all positions
	tx, err := db.Begin()
	if err != nil {
		common.Error("%v\n", err)
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf(
		"DELETE FROM %s WHERE castell_model_id = (SELECT id FROM %s WHERE uuid = ?)",
		CASTELLMEMBERPOSITIONSTABLE, CASTELLMODELSTABLE))
	defer stmt.Close()
	if err != nil {
		tx.Rollback()
		common.Error("%v\n", err)
		return err
	}
	_, err = stmt.Exec(
		c.UUID,
	)
	if err != nil {
		tx.Rollback()
		common.Error("%v\n", err)
		return err
	}
	// add all positions
	for _, member := range c.PositionMembers {
		stmt, err = tx.Prepare(fmt.Sprintf(
			"INSERT INTO %s "+
				"(castell_model_id, castell_position_id, member_id) VALUES ("+
				"(SELECT id FROM %s WHERE uuid = ?), "+
				"(SELECT id FROM %s WHERE name = ? AND column = ? AND cordon = ? AND part = ?), "+
				"(SELECT id FROM %s WHERE uuid = ?))",
			CASTELLMEMBERPOSITIONSTABLE, CASTELLMODELSTABLE, CASTELLPOSITIONSTABLE, MEMBERSTABLE))
		defer stmt.Close()
		if err != nil {
			tx.Rollback()
			common.Error("%v\n", err)
			return err
		}
		_, err = stmt.Exec(
			c.UUID,
			member.Position.Name,
			member.Position.Column,
			member.Position.Cordon,
			member.Position.Part,
			member.MemberUUID,
		)
		if err != nil {
			tx.Rollback()
			common.Error("%v\n", err)
			return err
		}
	}
	// update model fields
	stmt, err = tx.Prepare(fmt.Sprintf(
		"UPDATE %s SET name=?, castell_type_name=? WHERE uuid=?",
		CASTELLMODELSTABLE))
	defer stmt.Close()
	if err != nil {
		tx.Rollback()
		common.Error("%v\n", err)
		return err
	}
	_, err = stmt.Exec(
		c.Name,
		c.Type,
		c.UUID,
	)
	if err != nil {
		tx.Rollback()
		common.Error("%v\n", err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		common.Error("%v\n", err)
		tx.Rollback()
	}
	return err
}

func (c *CastellModel) GetAll() ([]CastellModel, error) {
	rows, err := db.Query(fmt.Sprintf(
		"SELECT uuid, name, castell_type_name FROM %s WHERE deleted=0",
		CASTELLMODELSTABLE))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer rows.Close()

	models := []CastellModel{}

	for rows.Next() {
		var c CastellModel
		if err = rows.Scan(&c.UUID, &c.Name, &c.Type); err != nil {
			return nil, err
		}
		models = append(models, c)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return models, nil
}

func (c *CastellModel) Get() error {
	stmt, err := db.Prepare(fmt.Sprintf(
		"SELECT model_name, model_type, position_in_castell_name, position_in_castell_column, position_in_castell_cordon, position_in_castell_part, member_uuid FROM %s WHERE model_uuid = ? AND model_deleted=0",
		CASTELLMODELSVIEW))
	if err != nil {
		common.Fatal(err.Error())
	}
	rows, err := stmt.Query(c.UUID)
	defer rows.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	for rows.Next() {
		var p CastellPositionMembers
		if err = rows.Scan(&c.Name, &c.Type, &p.Position.Name, &p.Position.Column, &p.Position.Cordon, &p.Position.Part, &p.MemberUUID); err != nil {
			return err
		}
		c.PositionMembers = append(c.PositionMembers, p)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return nil
}

func (c *CastellModel) Delete() error {
	stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s SET deleted=1 WHERE uuid=?",
		CASTELLMODELSTABLE))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	_, err = stmt.Exec(c.UUID)
	return err
}

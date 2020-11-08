package model

import (
	"errors"
	"fmt"
	"strings"

	"github.com/vilisseranen/castellers/common"
)

const MembersTable = "members"
const MembersCredentialsTable = "members_credentials"

const MemberTypeAdmin = "admin"
const MemberTypeMember = "member"

type Member struct {
	UUID          string   `json:"uuid"`
	FirstName     string   `json:"firstName"` // Encrypted
	LastName      string   `json:"lastName"`  // Encrypted
	Height        string   `json:"height"`    // Encrypted
	Weight        string   `json:"weight"`    // Encrypted
	Roles         []string `json:"roles"`     // Encrypted
	Extra         string   `json:"extra"`     // Encrypted
	Type          string   `json:"type"`      // Encrypted
	Email         string   `json:"email"`     // Encrypted
	Contact       string   `json:"contact"`   // Encrypted
	Code          string   `json:"-"`
	Activated     int      `json:"activated"`
	Subscribed    int      `json:"subscribed"`
	Deleted       int      `json:"-"`
	Language      string   `json:"language"`
	Participation string   `json:"participation"`
	Presence      string   `json:"presence"`
}

type Credentials struct {
	UUID           string
	Username       string
	Password       string
	PasswordHashed []byte
}

func (m *Member) CreateMember() error {
	tx, err := db.Begin()
	if err != nil {
		common.Error("%v\n", m)
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf(
		"INSERT INTO %s (uuid, firstName, lastName, height, weight, roles, extra, type, email, contact, code, language) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		MembersTable))
	if err != nil {
		common.Error("%v\n", m)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		stringOrNull(m.UUID),
		common.Encrypt(m.FirstName),
		common.Encrypt(m.LastName),
		common.Encrypt(m.Height),
		common.Encrypt(m.Weight),
		common.Encrypt(strings.Join(m.Roles, ",")),
		common.Encrypt(m.Extra),
		common.Encrypt(m.Type),
		common.Encrypt(m.Email),
		common.Encrypt(m.Contact),
		stringOrNull(m.Code),
		stringOrNull(m.Language))
	if err != nil {
		common.Error("%v\n", m)
		return err
	}
	err = tx.Commit()
	return err
}

func (m *Member) EditMember(callerType string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	switch callerType {
	case MemberTypeAdmin:
		stmt, err := tx.Prepare(fmt.Sprintf(
			"UPDATE %s SET firstName=?, lastName=?, height=?, weight=?, roles=?, extra=?, type=?, email=?, contact=?, language=?, subscribed=? WHERE uuid=?",
			MembersTable))
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(
			common.Encrypt(m.FirstName),
			common.Encrypt(m.LastName),
			common.Encrypt(m.Height),
			common.Encrypt(m.Weight),
			common.Encrypt(strings.Join(m.Roles, ",")),
			common.Encrypt(m.Extra),
			common.Encrypt(m.Type),
			common.Encrypt(m.Email),
			common.Encrypt(m.Contact),
			m.Language,
			m.Subscribed,
			stringOrNull(m.UUID))
	case MemberTypeMember:
		stmt, err := tx.Prepare(fmt.Sprintf(
			"UPDATE %s SET firstName=?, lastName=?, height=?, weight=?, type=?, email=?, contact=?, language=?, subscribed=? WHERE uuid=?",
			MembersTable))
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(
			common.Encrypt(m.FirstName),
			common.Encrypt(m.LastName),
			common.Encrypt(m.Height),
			common.Encrypt(m.Weight),
			common.Encrypt(m.Type),
			common.Encrypt(m.Email),
			common.Encrypt(m.Contact),
			stringOrNull(m.Language),
			m.Subscribed,
			stringOrNull(m.UUID))
	default:
		err = errors.New("")
	}
	if err != nil {
		common.Error("%v\n", m)
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
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	var rolesAsString string
	err = stmt.QueryRow(m.UUID).Scan(&m.FirstName, &m.LastName, &m.Height, &m.Weight, &rolesAsString, &m.Extra, &m.Type, &m.Email, &m.Contact, &m.Code, &m.Activated, &m.Subscribed, &m.Language)
	if err == nil {
		m.FirstName = common.Decrypt([]byte(m.FirstName))
		m.LastName = common.Decrypt([]byte(m.LastName))
		m.Height = common.Decrypt([]byte(m.Height))
		m.Weight = common.Decrypt([]byte(m.Weight))
		m.Roles = strings.Split(common.Decrypt([]byte(rolesAsString)), ",")
		m.Extra = common.Decrypt([]byte(m.Extra))
		m.Type = common.Decrypt([]byte(m.Type))
		m.Email = common.Decrypt([]byte(m.Email))
		m.Contact = common.Decrypt([]byte(m.Contact))
		m.sanitizeEmptyRoles()
	}
	return err
}

func (m *Member) GetAll() ([]Member, error) {
	rows, err := db.Query(fmt.Sprintf(
		"SELECT uuid, firstName, lastName, height, weight, roles, extra, type, email, contact, code, activated, subscribed, language FROM %s WHERE deleted=0",
		MembersTable))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer rows.Close()

	members := []Member{}

	for rows.Next() {
		var m Member
		var rolesAsString string
		if err = rows.Scan(&m.UUID, &m.FirstName, &m.LastName, &m.Height, &m.Weight, &rolesAsString, &m.Extra, &m.Type, &m.Email, &m.Contact, &m.Code, &m.Activated, &m.Subscribed, &m.Language); err != nil {
			return nil, err
		}
		m.FirstName = common.Decrypt([]byte(m.FirstName))
		m.LastName = common.Decrypt([]byte(m.LastName))
		m.Height = common.Decrypt([]byte(m.Height))
		m.Weight = common.Decrypt([]byte(m.Weight))
		m.Roles = strings.Split(common.Decrypt([]byte(rolesAsString)), ",")
		m.Extra = common.Decrypt([]byte(m.Extra))
		m.Type = common.Decrypt([]byte(m.Type))
		m.Email = common.Decrypt([]byte(m.Email))
		m.Contact = common.Decrypt([]byte(m.Contact))
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
		common.Fatal(err.Error())
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
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.UUID)
	return err
}

func (c *Credentials) CreateCredentials(username string, password []byte) error {
	stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO %s (uuid, username, password) VALUES (?, ?, ?)", MembersCredentialsTable))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(c.UUID, username, password)
	return err
}

func (c *Credentials) GetCredentials() error {
	stmt, err := db.Prepare(fmt.Sprintf(
		"SELECT uuid, password FROM %s WHERE username= ?",
		MembersCredentialsTable))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	common.Debug("username: %s", c.Username)
	err = stmt.QueryRow(c.Username).Scan(&c.UUID, &c.PasswordHashed)
	return err
}

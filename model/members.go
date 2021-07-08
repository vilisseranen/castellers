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

const MemberEmailNotFoundMessage = "No member found with this email"

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
	UUID           string `json:"-"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	PasswordHashed []byte `json:"-"`
}

func (m *Member) CreateMember() error {
	stmt, err := db.Prepare(fmt.Sprintf(
		"INSERT INTO %s (uuid, firstName, lastName, height, weight, roles, extra, type, email, contact, code, language) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		MembersTable))
	defer stmt.Close()
	if err != nil {
		common.Error(err.Error())
		common.Error("%v\n", m)
		return err
	}
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
		common.Error(err.Error())
		common.Error("%v\n", m)
		return err
	}
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
	defer stmt.Close()
	if err != nil {
		tx.Rollback()
		common.Error("%v\n")
		return err
	}
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
	if err != nil {
		tx.Rollback()
		common.Error("%v\n", m)
		return err
	}
	err = tx.Commit()
	if err != nil {
		common.Error("%v\n", err)
		tx.Rollback()
	}
	return err
}

func (m *Member) Get() error {
	stmt, err := db.Prepare(fmt.Sprintf(
		"SELECT firstName, lastName, height, weight, roles, extra, type, email, contact, code, activated, subscribed, language FROM %s WHERE uuid= ? AND deleted=0",
		MembersTable))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
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
	defer rows.Close()
	if err != nil {
		common.Fatal(err.Error())
		return nil, err
	}

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
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
		return err
	}

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
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
		return err
	}
	_, err = stmt.Exec(m.UUID)
	return err
}

func (c *Credentials) ResetCredentials(username string, password []byte) error {
	member := Member{UUID: c.UUID}
	err := member.Activate()
	if err != nil {
		common.Fatal(err.Error())
	}
	stmt, err := db.Prepare(fmt.Sprintf("DELETE FROM %s WHERE uuid = ?", MembersCredentialsTable))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
		return err
	}
	_, err = stmt.Exec(c.UUID)
	if err != nil {
		common.Fatal(err.Error())
		return err
	}
	stmt, err = db.Prepare(fmt.Sprintf("INSERT INTO %s (uuid, username, password) VALUES (?, ?, ?)", MembersCredentialsTable))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
		return err
	}

	_, err = stmt.Exec(c.UUID, username, password)
	return err
}

func (c *Credentials) GetCredentials() error {
	stmt, err := db.Prepare(fmt.Sprintf(
		"SELECT uuid, password FROM %s WHERE username= ?",
		MembersCredentialsTable))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	err = stmt.QueryRow(c.Username).Scan(&c.UUID, &c.PasswordHashed)
	return err
}

func (c *Credentials) GetCredentialsByUUID() error {
	stmt, err := db.Prepare(fmt.Sprintf(
		"SELECT username FROM %s WHERE uuid= ?",
		MembersCredentialsTable))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	err = stmt.QueryRow(c.UUID).Scan(&c.Username)
	return err
}

func (m *Member) GetByEmail() error {
	found := false
	members, err := m.GetAll()
	if err != nil {
		return err
	}
	for _, member := range members {
		if member.Email == m.Email {
			common.Debug("Found a member with email %s", m.Email)
			*m = member
			found = true
		}
	}
	if found == false {
		common.Debug("Email %s not found.", m.Email)
		return errors.New(MemberEmailNotFoundMessage)
	}
	return nil
}

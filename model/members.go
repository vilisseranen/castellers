package model

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/vilisseranen/castellers/common"
)

const (
	MEMBERSTABLE            = "members"
	MEMBERSCREDENTIALSTABLE = "members_credentials"

	MEMBERSTYPEADMIN   = "admin"
	MEMBERSTYPEREGULAR = "member"
	MEMBERSTYPEGUEST   = "guest"

	MEMBERSSTATUSCREATED   = "created"
	MEMBERSSTATUSACTIVATED = "active"
	MEMBERSSTATUSPAUSED    = "paused"
	MEMBERSSTATUSDELETED   = "deleted"
	MEMBERSSTATUSPURGED    = "purged"

	MEMBERSEMAILNOTFOUNDMESSAGE = "No member found with this email"
)

type Member struct {
	UUID          string   `json:"uuid"`
	FirstName     string   `json:"firstName"` // Encrypted
	LastName      string   `json:"lastName"`  // Encrypted
	Height        string   `json:"height"`    // Encrypted
	Weight        string   `json:"weight"`    // Encrypted
	Roles         []string `json:"roles"`     // Encrypted
	Extra         string   `json:"extra"`     // Encrypted
	Email         string   `json:"email"`     // Encrypted
	Contact       string   `json:"contact"`   // Encrypted
	Type          string   `json:"type"`
	Status        string   `json:"status"`
	Subscribed    int      `json:"subscribed"`
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

func (m *Member) CreateMember(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Member.CreateMember")
	defer span.End()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(
		"INSERT INTO %s (uuid, firstName, lastName, height, weight, roles, extra, type, email, contact, language) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		MEMBERSTABLE))
	defer stmt.Close()
	if err != nil {
		tx.Rollback()
		common.Error("Error: %v on member: %v", err.Error(), m)
		return err
	}
	_, err = stmt.ExecContext(
		ctx,
		stringOrNull(m.UUID),
		common.Encrypt(m.FirstName),
		common.Encrypt(m.LastName),
		common.Encrypt(m.Height),
		common.Encrypt(m.Weight),
		common.Encrypt(strings.Join(m.Roles, ",")),
		common.Encrypt(m.Extra),
		m.Type,
		common.Encrypt(m.Email),
		common.Encrypt(m.Contact),
		stringOrNull(m.Language))
	if err != nil {
		tx.Rollback()
		common.Error("Error: %v on member: %v", err.Error(), m)
		return err
	}
	stmt, err = tx.PrepareContext(ctx, fmt.Sprintf(
		"UPDATE %s SET status = '%s' WHERE uuid = ?",
		MEMBERSTABLE, MEMBERSSTATUSACTIVATED))
	defer stmt.Close()
	if err != nil {
		tx.Rollback()
		common.Error("Error: %v on member: %v", err.Error(), m)
		return err
	}
	_, err = stmt.ExecContext(
		ctx,
		stringOrNull(m.UUID))
	if err != nil {
		tx.Rollback()
		common.Error("Error: %v on member: %v", err.Error(), m)
		return err
	}
	err = tx.Commit()
	if err != nil {
		common.Error("%v\n", err)
		tx.Rollback()
	}
	return err
}

func (m *Member) EditMember(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Member.EditMember")
	defer span.End()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(
		"UPDATE %s SET firstName=?, lastName=?, height=?, weight=?, roles=?, extra=?, type=?, email=?, contact=?, language=?, subscribed=? WHERE uuid=?",
		MEMBERSTABLE))
	defer stmt.Close()
	if err != nil {
		tx.Rollback()
		common.Error("%v\n")
		return err
	}
	_, err = stmt.ExecContext(ctx,
		common.Encrypt(m.FirstName),
		common.Encrypt(m.LastName),
		common.Encrypt(m.Height),
		common.Encrypt(m.Weight),
		common.Encrypt(strings.Join(m.Roles, ",")),
		common.Encrypt(m.Extra),
		m.Type,
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

func (m *Member) Get(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Member.Get")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"SELECT firstName, lastName, height, weight, roles, extra, type, email, contact, status, subscribed, language FROM %s WHERE uuid= ? AND status != '%s'",
		MEMBERSTABLE, MEMBERSSTATUSDELETED))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	var rolesAsString string
	err = stmt.QueryRowContext(ctx, m.UUID).Scan(&m.FirstName, &m.LastName, &m.Height, &m.Weight, &rolesAsString, &m.Extra, &m.Type, &m.Email, &m.Contact, &m.Status, &m.Subscribed, &m.Language)
	if err == nil {
		m.FirstName = common.Decrypt([]byte(m.FirstName))
		m.LastName = common.Decrypt([]byte(m.LastName))
		m.Height = common.Decrypt([]byte(m.Height))
		m.Weight = common.Decrypt([]byte(m.Weight))
		m.Roles = strings.Split(common.Decrypt([]byte(rolesAsString)), ",")
		m.Extra = common.Decrypt([]byte(m.Extra))
		m.Email = common.Decrypt([]byte(m.Email))
		m.Contact = common.Decrypt([]byte(m.Contact))
		m.sanitizeEmptyRoles()
	}
	return err
}

func (m *Member) GetAll(ctx context.Context, memberStatusList, memberTypeList []string) ([]Member, error) {
	ctx, span := tracer.Start(ctx, "Member.GetAll")
	defer span.End()
	queryString := []string{fmt.Sprintf(
		"SELECT uuid, firstName, lastName, height, weight, roles, extra, type, email, contact, status, subscribed, language FROM %s",
		MEMBERSTABLE)}
	filters := []string{}
	statusFilters := []string{}
	typeFilters := []string{}
	queryValues := []interface{}{}

	// filter on status
	for _, status := range memberStatusList {
		if status != "" {
			statusFilters = append(statusFilters, "status = ?")
			queryValues = append(queryValues, status)
		}
	}
	if len(statusFilters) > 0 {
		filters = append(filters, fmt.Sprintf("( %s )", strings.Join(statusFilters, " OR ")))
	}

	// filter on type
	for _, mType := range memberTypeList {
		if mType != "" {
			typeFilters = append(typeFilters, "type = ?")
			queryValues = append(queryValues, mType)
		}
	}
	if len(typeFilters) > 0 {
		filters = append(filters, fmt.Sprintf("( %s )", strings.Join(typeFilters, " OR ")))
	}

	filters = append(filters, fmt.Sprintf("status NOT IN ('%s', '%s')", MEMBERSSTATUSDELETED, MEMBERSSTATUSPURGED))

	filter := strings.Join(filters, " AND ")
	queryString = compact(append(queryString, filter))
	query := strings.Join(queryString, " WHERE ")
	common.Debug("SQL query: %s; params(values=%v)", query, queryValues)
	rows, err := db.QueryContext(ctx, query, queryValues...)
	defer rows.Close()
	if err != nil {
		common.Fatal(err.Error())
		return nil, err
	}

	members := []Member{}

	for rows.Next() {
		var m Member
		var rolesAsString string
		if err = rows.Scan(&m.UUID, &m.FirstName, &m.LastName, &m.Height, &m.Weight, &rolesAsString, &m.Extra, &m.Type, &m.Email, &m.Contact, &m.Status, &m.Subscribed, &m.Language); err != nil {
			return nil, err
		}
		m.FirstName = common.Decrypt([]byte(m.FirstName))
		m.LastName = common.Decrypt([]byte(m.LastName))
		m.Height = common.Decrypt([]byte(m.Height))
		m.Weight = common.Decrypt([]byte(m.Weight))
		m.Roles = strings.Split(common.Decrypt([]byte(rolesAsString)), ",")
		m.Extra = common.Decrypt([]byte(m.Extra))
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

func (m *Member) DeleteMember(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Member.DeleteMember")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("UPDATE %s SET status='%s' WHERE uuid=?",
		MEMBERSTABLE, MEMBERSSTATUSDELETED))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
		return err
	}

	_, err = stmt.ExecContext(ctx, m.UUID)
	return err
}

func (m *Member) sanitizeEmptyRoles() {
	if len(m.Roles) == 1 && m.Roles[0] == "" {
		m.Roles = []string{}
	}
	return
}

func (m *Member) SetStatus(ctx context.Context, status string) error {
	ctx, span := tracer.Start(ctx, "Member.SetStatus")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("UPDATE %s SET status = ? WHERE uuid= ?", MEMBERSTABLE))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
		return err
	}
	_, err = stmt.ExecContext(ctx, status, m.UUID)
	return err
}

func (c *Credentials) ResetCredentials(ctx context.Context, username string, password []byte) error {
	ctx, span := tracer.Start(ctx, "Credentials.ResetCredentials")
	defer span.End()
	member := Member{UUID: c.UUID}
	err := member.SetStatus(ctx, MEMBERSSTATUSACTIVATED)
	if err != nil {
		common.Fatal(err.Error())
	}
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE uuid = ?", MEMBERSCREDENTIALSTABLE))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
		return err
	}
	_, err = stmt.ExecContext(ctx, c.UUID)
	if err != nil {
		common.Fatal(err.Error())
		return err
	}
	stmt, err = db.PrepareContext(ctx, fmt.Sprintf("INSERT INTO %s (uuid, username, password) VALUES (?, ?, ?)", MEMBERSCREDENTIALSTABLE))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
		return err
	}

	_, err = stmt.ExecContext(ctx, c.UUID, username, password)
	return err
}

func (c *Credentials) GetCredentials(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Credentials.GetCredentials")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"SELECT uuid, password FROM %s WHERE username= ?",
		MEMBERSCREDENTIALSTABLE))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	err = stmt.QueryRowContext(ctx, c.Username).Scan(&c.UUID, &c.PasswordHashed)
	return err
}

func (c *Credentials) GetCredentialsByUUID(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Credentials.GetCredentialsByUUID")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"SELECT username FROM %s WHERE uuid= ?",
		MEMBERSCREDENTIALSTABLE))
	defer stmt.Close()
	if err != nil {
		common.Fatal(err.Error())
	}
	err = stmt.QueryRowContext(ctx, c.UUID).Scan(&c.Username)
	return err
}

func (m *Member) GetByEmail(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Member.GetByEmail")
	defer span.End()
	found := false
	members, err := m.GetAll(ctx, []string{}, []string{})
	if err != nil {
		return err
	}
	for _, member := range members {
		if strings.ToLower(member.Email) == strings.ToLower(m.Email) {
			common.Debug("Found a member with email %s", m.Email)
			*m = member
			found = true
		}
	}
	if found == false {
		common.Debug("Email %s not found.", m.Email)
		return errors.New(MEMBERSEMAILNOTFOUNDMESSAGE)
	}
	return nil
}

func compact(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

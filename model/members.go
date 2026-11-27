package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vilisseranen/castellers/common"
)

const (
	MEMBERSTABLE            = "members"
	MEMBERSCREDENTIALSTABLE = "members_credentials"
	MEMBERSDEPENDANTSTABLE  = "members_dependent"

	MEMBERSTYPEADMIN   = "admin"
	MEMBERSTYPEREGULAR = "member"
	MEMBERSTYPECANALLA = "canalla"
	MEMBERSTYPEGUEST   = "guest"

	MEMBERSSTATUSCREATED   = "created"
	MEMBERSSTATUSACTIVATED = "active"
	MEMBERSSTATUSPAUSED    = "paused"
	MEMBERSSTATUSDELETED   = "deleted"
	MEMBERSSTATUSPURGED    = "purged"

	MEMBERSEMAILNOTFOUNDMESSAGE = "no member found with this email"
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
	// CreatedAt is when the member was added (Unix seconds). It is 0 when it is
	// not known, and it is only filled in the list of members.
	CreatedAt int64 `json:"createdAt"`
	// Musician is 1 when the member plays with the musicians of the colla. It
	// does not depend on the type: a casteller can be a musician too. Only
	// admins can change it.
	Musician int `json:"musician"`
	// ShirtSize is the size of the shirt of the colla the member has: an even
	// number from MINSHIRTSIZE to MAXSHIRTSIZE, or 0 when it is not known or
	// they have no shirt. ShirtSince is the day they got it (YYYY-MM-DD), or
	// empty. Only admins can change them.
	ShirtSize  int    `json:"shirtSize"`
	ShirtSince string `json:"shirtSince"`
	// DeletedAt is when the member was deleted (Unix seconds) and PurgeAt when
	// their personal data will be erased. Both are 0 for a member who is not
	// deleted, and they are only filled in the list of members.
	DeletedAt int64 `json:"deletedAt"`
	PurgeAt   int64 `json:"purgeAt"`
	// DeletionAt is when the member will be deleted automatically, because
	// they have not come for a long time or never activated their account
	// (Unix seconds). It is 0 when no deletion is planned, and it is only
	// filled in the list of members.
	DeletionAt int64 `json:"deletionAt"`
}

// ErrMemberNotDeleted is returned when restoring a member who is not deleted,
// or whose personal data is already erased.
var ErrMemberNotDeleted = errors.New("member is not deleted")

// PurgeDelay is how long a deleted member is kept before their personal data
// is erased, in seconds. An admin can restore the member during that time.
func PurgeDelay() int64 {
	return int64(common.GetConfigInt("purge_delay_days")) * 24 * 3600
}

// DeletionNoticeDelay is the time between the notice sent to an inactive
// member and their automatic deletion, in seconds.
func DeletionNoticeDelay() int64 {
	return int64(common.GetConfigInt("deletion_notice_days")) * 24 * 3600
}

// String keeps personal data out of logs and error messages: formatting a
// Member with %v or %s only shows its UUID.
func (m Member) String() string {
	return "Member{" + m.UUID + "}"
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
		"INSERT INTO %s (uuid, firstName, lastName, height, weight, roles, extra, type, email, contact, language, musician, shirt_size, shirt_since, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		MEMBERSTABLE))
	if err != nil {
		tx.Rollback()
		common.Error("Error: %v on member %s", err.Error(), m.UUID)
		return err
	}
	defer stmt.Close()
	createdAt := time.Now().Unix()
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
		stringOrNull(m.Language),
		musicianFlag(m.Musician),
		intOrNull(m.ShirtSize),
		stringOrNull(m.ShirtSince),
		createdAt)
	if err != nil {
		tx.Rollback()
		common.Error("Error: %v on member %s", err.Error(), m.UUID)
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
		"UPDATE %s SET firstName=?, lastName=?, height=?, weight=?, roles=?, extra=?, type=?, email=?, contact=?, language=?, subscribed=?, musician=?, shirt_size=?, shirt_since=? WHERE uuid=?",
		MEMBERSTABLE))
	if err != nil {
		tx.Rollback()
		common.Error("Error editing member %s: %v", m.UUID, err)
		return err
	}
	defer stmt.Close()
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
		musicianFlag(m.Musician),
		intOrNull(m.ShirtSize),
		stringOrNull(m.ShirtSince),
		stringOrNull(m.UUID))
	if err != nil {
		tx.Rollback()
		common.Error("Error on member %s: %v", m.UUID, err)
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
		"SELECT firstName, lastName, height, weight, roles, extra, type, email, contact, status, subscribed, language, musician, shirt_size, shirt_since FROM %s WHERE uuid= ? AND status NOT IN (?, ?)",
		MEMBERSTABLE))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	var rolesAsString string
	var shirtSize sql.NullInt64
	var shirtSince sql.NullString
	err = stmt.QueryRowContext(ctx, m.UUID, MEMBERSSTATUSDELETED, MEMBERSSTATUSPURGED).Scan(&m.FirstName, &m.LastName, &m.Height, &m.Weight, &rolesAsString, &m.Extra, &m.Type, &m.Email, &m.Contact, &m.Status, &m.Subscribed, &m.Language, &m.Musician, &shirtSize, &shirtSince)
	if err == nil {
		m.ShirtSize = int(shirtSize.Int64)
		m.ShirtSince = nullToEmptyString(shirtSince)
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

// GetAll lists the members with one of the statuses and one of the types given;
// an empty list means any. Deleted and purged members are only listed when
// their status is asked for explicitly. A purged member has no personal data
// left: all their fields are empty.
func (m *Member) GetAll(ctx context.Context, memberStatusList, memberTypeList []string) ([]Member, error) {
	ctx, span := tracer.Start(ctx, "Member.GetAll")
	defer span.End()
	queryString := []string{fmt.Sprintf(
		"SELECT uuid, firstName, lastName, height, weight, roles, extra, type, email, contact, status, subscribed, language, musician, shirt_size, shirt_since, created_at, deleted_at, deletion_notice_at FROM %s",
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

	// Former members are left out, unless their status was asked for
	for _, status := range []string{MEMBERSSTATUSDELETED, MEMBERSSTATUSPURGED} {
		if !common.StringInSlice(status, memberStatusList) {
			filters = append(filters, "status != ?")
			queryValues = append(queryValues, status)
		}
	}

	filter := strings.Join(filters, " AND ")
	queryString = compact(append(queryString, filter))
	query := strings.Join(queryString, " WHERE ")
	common.Debug("SQL query: %s; params(values=%v)", query, queryValues)
	rows, err := db.QueryContext(ctx, query, queryValues...)
	if err != nil {
		common.Fatal(err.Error())
		return nil, err
	}
	defer rows.Close()

	members := []Member{}

	for rows.Next() {
		var m Member
		var rolesAsString string
		var shirtSize sql.NullInt64
		var shirtSince sql.NullString
		var deletedAt, deletionNoticeAt sql.NullInt64
		if err = rows.Scan(&m.UUID, &m.FirstName, &m.LastName, &m.Height, &m.Weight, &rolesAsString, &m.Extra, &m.Type, &m.Email, &m.Contact, &m.Status, &m.Subscribed, &m.Language, &m.Musician, &shirtSize, &shirtSince, &m.CreatedAt, &deletedAt, &deletionNoticeAt); err != nil {
			return nil, err
		}
		if m.Status == MEMBERSSTATUSPURGED {
			// Nothing is left of a purged member. Their row may also have been
			// scrubbed outside of the application: do not try to decrypt it.
			members = append(members, Member{UUID: m.UUID, Type: m.Type, Status: m.Status, Language: m.Language, Roles: []string{}, CreatedAt: m.CreatedAt, DeletedAt: deletedAt.Int64})
			continue
		}
		if m.Status == MEMBERSSTATUSDELETED && deletedAt.Int64 > 0 {
			m.DeletedAt = deletedAt.Int64
			m.PurgeAt = deletedAt.Int64 + PurgeDelay()
		}
		if m.Status != MEMBERSSTATUSDELETED && deletionNoticeAt.Int64 > 0 {
			m.DeletionAt = deletionNoticeAt.Int64 + DeletionNoticeDelay()
		}
		m.ShirtSize = int(shirtSize.Int64)
		m.ShirtSince = nullToEmptyString(shirtSince)
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

// DeleteMember is a soft delete: the member disappears from the application,
// but their data is kept until the purge (see PurgeDelay), so that an admin
// can restore them.
func (m *Member) DeleteMember(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Member.DeleteMember")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("UPDATE %s SET status=?, deleted_at=?, deletion_notice_at=NULL, admin_notified_at=NULL WHERE uuid=?",
		MEMBERSTABLE))
	if err != nil {
		common.Fatal(err.Error())
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, MEMBERSSTATUSDELETED, time.Now().Unix(), m.UUID)
	return err
}

// GetDeleted loads a deleted member who is not purged yet. Member.Get does not
// see them. It returns sql.ErrNoRows for anybody else.
func (m *Member) GetDeleted(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Member.GetDeleted")
	defer span.End()
	members, err := m.GetAll(ctx, []string{MEMBERSSTATUSDELETED}, []string{})
	if err != nil {
		return err
	}
	for _, member := range members {
		if member.UUID == m.UUID {
			*m = member
			return nil
		}
	}
	return sql.ErrNoRows
}

// RestoreMember brings back a deleted member, as long as they are not purged.
// A member who had chosen a password, a guest or a canalla comes back active;
// a member who never chose a password comes back as created. It returns
// ErrMemberNotDeleted when there is nobody to restore.
func (m *Member) RestoreMember(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Member.RestoreMember")
	defer span.End()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var memberType string
	var credentials int
	err = tx.QueryRowContext(ctx, fmt.Sprintf(
		"SELECT type, (SELECT COUNT(*) FROM %s c WHERE c.uuid = m.uuid) FROM %s m WHERE m.uuid = ? AND m.status = ?",
		MEMBERSCREDENTIALSTABLE, MEMBERSTABLE), m.UUID, MEMBERSSTATUSDELETED).Scan(&memberType, &credentials)
	if err == sql.ErrNoRows {
		return ErrMemberNotDeleted
	}
	if err != nil {
		return err
	}
	status := MEMBERSSTATUSCREATED
	if credentials > 0 || memberType == MEMBERSTYPEGUEST || memberType == MEMBERSTYPECANALLA {
		status = MEMBERSSTATUSACTIVATED
	}
	// last_activity_date: the inactivity scan must not pause them right away
	if _, err = tx.ExecContext(ctx, fmt.Sprintf(
		"UPDATE %s SET status = ?, deleted_at = NULL, admin_notified_at = NULL, last_activity_date = ? WHERE uuid = ? AND status = ?",
		MEMBERSTABLE), status, time.Now().Unix(), m.UUID, MEMBERSSTATUSDELETED); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	m.Status = status
	return nil
}

// GetMembersToPurge returns the UUID of the members deleted at or before the
// given time (Unix seconds) and not purged yet.
func GetMembersToPurge(ctx context.Context, deletedBefore int64) ([]string, error) {
	ctx, span := tracer.Start(ctx, "GetMembersToPurge")
	defer span.End()
	rows, err := db.QueryContext(ctx, fmt.Sprintf(
		"SELECT uuid FROM %s WHERE status = ? AND deleted_at IS NOT NULL AND deleted_at <= ?",
		MEMBERSTABLE), MEMBERSSTATUSDELETED, deletedBefore)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	uuids := []string{}
	for rows.Next() {
		var uuid string
		if err := rows.Scan(&uuid); err != nil {
			return nil, err
		}
		uuids = append(uuids, uuid)
	}
	return uuids, rows.Err()
}

// PurgeMember erases the personal data of a deleted member, for good: name,
// email, contact, height, weight, roles, notes, shirt, credentials, links with
// dependents and the notes written about their dues. The row stays, with its
// UUID and its type, so that the past events, castells, dues and badges still
// count somebody: an anonymous former member. It returns false when the
// member is not deleted (nothing is erased then).
func (m *Member) PurgeMember(ctx context.Context) (bool, error) {
	ctx, span := tracer.Start(ctx, "Member.PurgeMember")
	defer span.End()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, fmt.Sprintf(
		"UPDATE %s SET firstName=?, lastName=?, height=?, weight=?, roles=?, extra=?, email=?, contact=?, subscribed=0, musician=0, shirt_size=NULL, shirt_since=NULL, status=? WHERE uuid=? AND status=?",
		MEMBERSTABLE),
		common.Encrypt(""),
		common.Encrypt(""),
		common.Encrypt(""),
		common.Encrypt(""),
		common.Encrypt(""),
		common.Encrypt(""),
		common.Encrypt(""),
		common.Encrypt(""),
		MEMBERSSTATUSPURGED,
		m.UUID,
		MEMBERSSTATUSDELETED)
	if err != nil {
		return false, err
	}
	purged, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if purged == 0 {
		return false, nil
	}
	if _, err = tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE uuid = ?", MEMBERSCREDENTIALSTABLE), m.UUID); err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, fmt.Sprintf(
		"DELETE FROM %s WHERE responsible_uuid = ? OR dependent_uuid = ?", MEMBERSDEPENDANTSTABLE), m.UUID, m.UUID); err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET note = NULL WHERE member_uuid = ?", DUES_LOG_TABLE), m.UUID); err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	m.Status = MEMBERSSTATUSPURGED
	return true, nil
}

// MemberActivity is what the inactivity scan needs to know about a member.
type MemberActivity struct {
	UUID   string
	Type   string
	Status string
	// Unix seconds, 0 when not known or never
	CreatedAt         int64
	LastActivityDate  int64
	LastParticipation int64 // start of the last event they answered or came to
	DeletionNoticeAt  int64 // when they were told of their deletion, 0 if not
}

// GetMembersActivity returns the activity of all the members who are not
// deleted. An answer to an event, whatever it is, shows that somebody is still
// around; so does being marked present.
func GetMembersActivity(ctx context.Context) ([]MemberActivity, error) {
	ctx, span := tracer.Start(ctx, "GetMembersActivity")
	defer span.End()
	rows, err := db.QueryContext(ctx, fmt.Sprintf(
		"SELECT m.uuid, m.type, m.status, m.created_at, m.last_activity_date, m.deletion_notice_at, "+
			"(SELECT MAX(e.startDate) FROM %s e JOIN %s p ON e.uuid = p.event_uuid "+
			"WHERE p.member_uuid = m.uuid AND (COALESCE(p.answer, '') != '' OR p.presence = ?) AND e.deleted = 0 AND e.cancelled = 0) "+
			"FROM %s m WHERE m.status NOT IN (?, ?)",
		EVENTS_TABLE, PARTICIPATION_TABLE, MEMBERSTABLE),
		common.AnswerYes, MEMBERSSTATUSDELETED, MEMBERSSTATUSPURGED)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	activities := []MemberActivity{}
	for rows.Next() {
		var a MemberActivity
		var noticeAt, lastParticipation sql.NullInt64
		if err := rows.Scan(&a.UUID, &a.Type, &a.Status, &a.CreatedAt, &a.LastActivityDate, &noticeAt, &lastParticipation); err != nil {
			return nil, err
		}
		a.DeletionNoticeAt = noticeAt.Int64
		a.LastParticipation = lastParticipation.Int64
		activities = append(activities, a)
	}
	return activities, rows.Err()
}

// SetDeletionNotice notes when a member was told of their automatic deletion.
// The admins are not told yet.
func (m *Member) SetDeletionNotice(ctx context.Context, at int64) error {
	ctx, span := tracer.Start(ctx, "Member.SetDeletionNotice")
	defer span.End()
	_, err := db.ExecContext(ctx, fmt.Sprintf(
		"UPDATE %s SET deletion_notice_at = ?, admin_notified_at = NULL WHERE uuid = ?", MEMBERSTABLE), at, m.UUID)
	return err
}

// ClearDeletionNotice cancels the automatic deletion of a member.
func (m *Member) ClearDeletionNotice(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Member.ClearDeletionNotice")
	defer span.End()
	_, err := db.ExecContext(ctx, fmt.Sprintf(
		"UPDATE %s SET deletion_notice_at = NULL, admin_notified_at = NULL WHERE uuid = ?", MEMBERSTABLE), m.UUID)
	return err
}

// KeepMember restarts the inactivity delay of a member and cancels their
// automatic deletion, if one was planned. Their status does not change.
func (m *Member) KeepMember(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Member.KeepMember")
	defer span.End()
	_, err := db.ExecContext(ctx, fmt.Sprintf(
		"UPDATE %s SET last_activity_date = ?, deletion_notice_at = NULL, admin_notified_at = NULL WHERE uuid = ? AND status NOT IN (?, ?)",
		MEMBERSTABLE), time.Now().Unix(), m.UUID, MEMBERSSTATUSDELETED, MEMBERSSTATUSPURGED)
	return err
}

// KeepMemberIfNoticed is KeepMember for a member who was told of their
// automatic deletion, and does nothing for the others: a member who logs in
// after the notice wants to stay.
func (m *Member) KeepMemberIfNoticed(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Member.KeepMemberIfNoticed")
	defer span.End()
	_, err := db.ExecContext(ctx, fmt.Sprintf(
		"UPDATE %s SET last_activity_date = ?, deletion_notice_at = NULL, admin_notified_at = NULL WHERE uuid = ? AND deletion_notice_at IS NOT NULL AND status NOT IN (?, ?)",
		MEMBERSTABLE), time.Now().Unix(), m.UUID, MEMBERSSTATUSDELETED, MEMBERSSTATUSPURGED)
	return err
}

// GetResponsibles returns the members who answer for this one (the parents of
// a canalla, usually), without those who are deleted.
func (m *Member) GetResponsibles(ctx context.Context) ([]Member, error) {
	ctx, span := tracer.Start(ctx, "Member.GetResponsibles")
	defer span.End()
	rows, err := db.QueryContext(ctx, fmt.Sprintf(
		"SELECT responsible_uuid FROM %s WHERE dependent_uuid = ?", MEMBERSDEPENDANTSTABLE), m.UUID)
	if err != nil {
		return nil, err
	}
	uuids := []string{}
	for rows.Next() {
		var uuid string
		if err := rows.Scan(&uuid); err != nil {
			rows.Close()
			return nil, err
		}
		uuids = append(uuids, uuid)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	responsibles := []Member{}
	for _, uuid := range uuids {
		responsible := Member{UUID: uuid}
		if err := responsible.Get(ctx); err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, err
		}
		responsibles = append(responsibles, responsible)
	}
	return responsibles, nil
}

// AdminNotice is a member the admins must be told about: somebody who will be
// deleted automatically, or a deleted member whose data will be erased soon.
type AdminNotice struct {
	UUID string
	Name string
	Type string
	Date int64 // day of the deletion, or of the purge (Unix seconds)
}

// GetAdminNotices lists what the admins were not told yet: the members with a
// planned deletion, and the deleted members whose purge comes before the given
// time.
func GetAdminNotices(ctx context.Context, purgeBefore int64) (deletions []AdminNotice, purges []AdminNotice, err error) {
	ctx, span := tracer.Start(ctx, "GetAdminNotices")
	defer span.End()
	rows, err := db.QueryContext(ctx, fmt.Sprintf(
		"SELECT uuid, firstName, lastName, type, status, deletion_notice_at, deleted_at FROM %s "+
			"WHERE admin_notified_at IS NULL AND ("+
			"(status NOT IN (?, ?) AND deletion_notice_at IS NOT NULL) OR "+
			"(status = ? AND deleted_at IS NOT NULL AND deleted_at + ? <= ?)) "+
			"ORDER BY id",
		MEMBERSTABLE),
		MEMBERSSTATUSDELETED, MEMBERSSTATUSPURGED, MEMBERSSTATUSDELETED, PurgeDelay(), purgeBefore)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	deletions = []AdminNotice{}
	purges = []AdminNotice{}
	for rows.Next() {
		var n AdminNotice
		var firstName, lastName []byte
		var status string
		var noticeAt, deletedAt sql.NullInt64
		if err := rows.Scan(&n.UUID, &firstName, &lastName, &n.Type, &status, &noticeAt, &deletedAt); err != nil {
			return nil, nil, err
		}
		n.Name = strings.TrimSpace(decryptOrEmpty(firstName) + " " + decryptOrEmpty(lastName))
		if status == MEMBERSSTATUSDELETED {
			n.Date = deletedAt.Int64 + PurgeDelay()
			purges = append(purges, n)
		} else {
			n.Date = noticeAt.Int64 + DeletionNoticeDelay()
			deletions = append(deletions, n)
		}
	}
	return deletions, purges, rows.Err()
}

// SetAdminNotified notes that the admins were told about these members.
func SetAdminNotified(ctx context.Context, uuids []string) error {
	ctx, span := tracer.Start(ctx, "SetAdminNotified")
	defer span.End()
	now := time.Now().Unix()
	for _, uuid := range uuids {
		if _, err := db.ExecContext(ctx, fmt.Sprintf(
			"UPDATE %s SET admin_notified_at = ? WHERE uuid = ?", MEMBERSTABLE), now, uuid); err != nil {
			return err
		}
	}
	return nil
}

func (m *Member) sanitizeEmptyRoles() {
	if len(m.Roles) == 1 && m.Roles[0] == "" {
		m.Roles = []string{}
	}
}

func (m *Member) SetStatus(ctx context.Context, status string) error {
	ctx, span := tracer.Start(ctx, "Member.SetStatus")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("UPDATE %s SET status = ? WHERE uuid= ?", MEMBERSTABLE))
	if err != nil {
		common.Fatal(err.Error())
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, status, m.UUID)
	return err
}

// SetStatusManual sets the status following an explicit admin action.
// Reactivating a member (status active) resets the inactivity counter by
// stamping last_activity_date, so pauseAbsentMembers treats the manual
// reactivation like a recent participation.
func (m *Member) SetStatusManual(ctx context.Context, status string) error {
	ctx, span := tracer.Start(ctx, "Member.SetStatusManual")
	defer span.End()
	var query string
	args := []interface{}{}
	if status == MEMBERSSTATUSACTIVATED {
		query = fmt.Sprintf("UPDATE %s SET status = ?, last_activity_date = ? WHERE uuid = ?", MEMBERSTABLE)
		args = append(args, status, time.Now().Unix(), m.UUID)
	} else {
		query = fmt.Sprintf("UPDATE %s SET status = ? WHERE uuid = ?", MEMBERSTABLE)
		args = append(args, status, m.UUID)
	}
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		common.Fatal(err.Error())
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, args...)
	return err
}

// GetLastActivityDate returns the last_activity_date stamp (Unix seconds, 0 if never set).
func (m *Member) GetLastActivityDate(ctx context.Context) (int64, error) {
	ctx, span := tracer.Start(ctx, "Member.GetLastActivityDate")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf("SELECT last_activity_date FROM %s WHERE uuid = ?", MEMBERSTABLE))
	if err != nil {
		common.Fatal(err.Error())
		return 0, err
	}
	defer stmt.Close()
	var lastActivityDate int64
	err = stmt.QueryRowContext(ctx, m.UUID).Scan(&lastActivityDate)
	return lastActivityDate, err
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
	if err != nil {
		common.Fatal(err.Error())
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, c.UUID)
	if err != nil {
		common.Fatal(err.Error())
		return err
	}
	stmt, err = db.PrepareContext(ctx, fmt.Sprintf("INSERT INTO %s (uuid, username, password) VALUES (?, ?, ?)", MEMBERSCREDENTIALSTABLE))
	if err != nil {
		common.Fatal(err.Error())
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, c.UUID, username, password)
	return err
}

func (c *Credentials) GetCredentials(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Credentials.GetCredentials")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"SELECT uuid, password FROM %s WHERE username= ?",
		MEMBERSCREDENTIALSTABLE))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, c.Username).Scan(&c.UUID, &c.PasswordHashed)
	return err
}

func (c *Credentials) GetCredentialsByUUID(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Credentials.GetCredentialsByUUID")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"SELECT username FROM %s WHERE uuid= ?",
		MEMBERSCREDENTIALSTABLE))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
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
		if strings.EqualFold(member.Email, m.Email) {
			common.Debug("Found member %s by email", member.UUID)
			*m = member
			found = true
		}
	}
	if !found {
		common.Debug("Email not found")
		return errors.New(MEMBERSEMAILNOTFOUNDMESSAGE)
	}
	return nil
}

func (m *Member) GetDependents(ctx context.Context) ([]Member, error) {
	ctx, span := tracer.Start(ctx, "Member.GetDependent")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"SELECT dependent_uuid FROM %s WHERE responsible_uuid = ?", MEMBERSDEPENDANTSTABLE,
	))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, m.UUID)
	if err != nil {
		common.Fatal(err.Error())
		return nil, err
	}
	defer rows.Close()
	var dependents []Member
	for rows.Next() {
		var dependent_uuid string
		var dependent Member
		if err = rows.Scan(&dependent_uuid); err != nil {
			return nil, err
		}
		common.Debug("Found dependent with UUID %s for responsible with UUID %s", dependent_uuid, m.UUID)
		dependent = Member{UUID: dependent_uuid}
		if err = dependent.Get(ctx); err != nil {
			if err == sql.ErrNoRows {
				// A deleted dependent: the link is kept in case they are restored
				continue
			}
			return nil, err
		}
		dependents = append(dependents, dependent)
	}
	return dependents, nil
}

func (m *Member) AddDependent(ctx context.Context, dependent *Member) error {
	ctx, span := tracer.Start(ctx, "Member.AddDependent")
	defer span.End()

	if err := dependent.Get(ctx); err != nil {
		return err
	}

	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"INSERT INTO %s(responsible_uuid, dependent_uuid) VALUES(?, ?) ON CONFLICT DO NOTHING", MEMBERSDEPENDANTSTABLE,
	))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, m.UUID, dependent.UUID)
	return err
}

func (m *Member) RemoveDependent(ctx context.Context, dependent *Member) error {
	ctx, span := tracer.Start(ctx, "Member.RemoveDependent")
	defer span.End()

	if err := dependent.Get(ctx); err != nil {
		return err
	}

	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"DELETE FROM %s WHERE responsible_uuid= ? AND dependent_uuid= ?", MEMBERSDEPENDANTSTABLE,
	))
	if err != nil {
		common.Fatal(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, m.UUID, dependent.UUID)
	return err
}

// musicianFlag keeps the musician column to 0 or 1, whatever was sent.
func musicianFlag(musician int) int {
	if musician != 0 {
		return 1
	}
	return 0
}

// intOrNull stores 0 as NULL: the value is not known.
func intOrNull(i int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: i != 0}
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

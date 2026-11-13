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
	DUES_COLLECTIONS_TABLE = "dues_collections"
	DUES_LOG_TABLE         = "dues_log"
	DUES_SETTINGS_TABLE    = "dues_settings"
)

// Kinds of collection. A season holds at most one of each.
const (
	DuesKindCercle = "cercle"
	DuesKindWinter = "winter"
	DuesKindSpring = "spring"
	DuesKindSummer = "summer"
)

// Actions written in the journal. The last action of a member for a collection
// is the state of that member for that collection.
const (
	DuesActionPaid   = "paid"
	DuesActionUnpaid = "unpaid"
	DuesActionExempt = "exempt"
)

// DuesNoteMaxLength is the maximum length of the note of a journal entry.
const DuesNoteMaxLength = 500

// DuesLogDefaultLimit and DuesLogMaxLimit bound a page of the journal.
const (
	DuesLogDefaultLimit = 100
	DuesLogMaxLimit     = 500
)

var ErrDuesCollectionNotFound = errors.New("dues collection not found")

// duesKindOrder sorts the collections in the order of the season.
const duesKindOrder = "CASE kind WHEN 'cercle' THEN 0 WHEN 'winter' THEN 1 WHEN 'spring' THEN 2 WHEN 'summer' THEN 3 ELSE 4 END"

// DuesCollection is one collection (cueillette) of a season. Season is the
// year in which the season starts: season 2026 runs from October 1st 2026 to
// September 30th 2027.
type DuesCollection struct {
	UUID   string `json:"uuid"`
	Season int    `json:"season"`
	Kind   string `json:"kind"`
	// Amount is in cents.
	Amount  int   `json:"amount"`
	DueDate int64 `json:"dueDate"`
}

// DuesState is the current state of a member for a collection.
type DuesState struct {
	MemberUUID     string `json:"memberUuid"`
	CollectionUUID string `json:"collectionUuid"`
	Status         string `json:"status"`
	UpdatedAt      int64  `json:"updatedAt"`
}

// DuesLogEntry is one line of the journal.
type DuesLogEntry struct {
	ID             int    `json:"id"`
	MemberUUID     string `json:"memberUuid"`
	MemberName     string `json:"memberName"`
	CollectionUUID string `json:"collectionUuid"`
	Season         int    `json:"season"`
	Kind           string `json:"kind"`
	Action         string `json:"action"`
	ActorUUID      string `json:"actorUuid"`
	ActorName      string `json:"actorName"`
	CreatedAt      int64  `json:"createdAt"`
	Note           string `json:"note"`
}

// DuesLogFilter narrows the journal. Empty fields are ignored.
type DuesLogFilter struct {
	MemberUUID     string
	CollectionUUID string
	ActorUUID      string
	Season         int
	Limit          int
	Offset         int
}

func ValidDuesKind(kind string) bool {
	return common.StringInSlice(kind, []string{DuesKindCercle, DuesKindWinter, DuesKindSpring, DuesKindSummer})
}

func ValidDuesAction(action string) bool {
	return common.StringInSlice(action, []string{DuesActionPaid, DuesActionUnpaid, DuesActionExempt})
}

// GetDuesCollections returns the collections of a season, or of every season
// when season is 0. Most recent season first, then in the order of the season.
func GetDuesCollections(ctx context.Context, season int) ([]DuesCollection, error) {
	ctx, span := tracer.Start(ctx, "GetDuesCollections")
	defer span.End()

	query := fmt.Sprintf("SELECT uuid, season, kind, amount, due_date FROM %s", DUES_COLLECTIONS_TABLE)
	args := []interface{}{}
	if season != 0 {
		query += " WHERE season = ?"
		args = append(args, season)
	}
	query += " ORDER BY season DESC, " + duesKindOrder
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	collections := []DuesCollection{}
	for rows.Next() {
		var c DuesCollection
		if err := rows.Scan(&c.UUID, &c.Season, &c.Kind, &c.Amount, &c.DueDate); err != nil {
			return nil, err
		}
		collections = append(collections, c)
	}
	return collections, rows.Err()
}

// SetDuesCollections creates or updates collections of a season. A kind that
// is not given is left untouched: collections are never removed, since the
// journal refers to them.
func SetDuesCollections(ctx context.Context, season int, collections []DuesCollection) error {
	ctx, span := tracer.Start(ctx, "SetDuesCollections")
	defer span.End()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(
		"INSERT INTO %s (uuid, season, kind, amount, due_date) VALUES (?, ?, ?, ?, ?) "+
			"ON CONFLICT(season, kind) DO UPDATE SET amount = excluded.amount, due_date = excluded.due_date",
		DUES_COLLECTIONS_TABLE))
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, c := range collections {
		if _, err := stmt.ExecContext(ctx, common.GenerateUUID(), season, c.Kind, c.Amount, c.DueDate); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// GetDuesStates returns the current state of each member for each collection,
// for the pairs that have at least one journal entry. A pair without entry is
// unpaid. season 0 and an empty memberUUID mean "all".
func GetDuesStates(ctx context.Context, season int, memberUUID string) ([]DuesState, error) {
	ctx, span := tracer.Start(ctx, "GetDuesStates")
	defer span.End()

	query := fmt.Sprintf(
		"SELECT l.member_uuid, l.collection_uuid, l.action, l.created_at FROM %s l "+
			"JOIN %s c ON c.uuid = l.collection_uuid "+
			"WHERE l.id IN (SELECT MAX(id) FROM %s GROUP BY member_uuid, collection_uuid)",
		DUES_LOG_TABLE, DUES_COLLECTIONS_TABLE, DUES_LOG_TABLE)
	args := []interface{}{}
	if season != 0 {
		query += " AND c.season = ?"
		args = append(args, season)
	}
	if memberUUID != "" {
		query += " AND l.member_uuid = ?"
		args = append(args, memberUUID)
	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := []DuesState{}
	for rows.Next() {
		var s DuesState
		if err := rows.Scan(&s.MemberUUID, &s.CollectionUUID, &s.Status, &s.UpdatedAt); err != nil {
			return nil, err
		}
		states = append(states, s)
	}
	return states, rows.Err()
}

// RecordDues writes one journal entry per collection for a member, all or
// nothing. A collection that is already in the requested state gets no entry,
// so repeating a request does not fill the journal. It returns the number of
// entries written, or ErrDuesCollectionNotFound if a collection does not exist.
func RecordDues(ctx context.Context, memberUUID string, collectionUUIDs []string, action, actorUUID, note string) (int, error) {
	ctx, span := tracer.Start(ctx, "RecordDues")
	defer span.End()

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	existsStmt, err := tx.PrepareContext(ctx, fmt.Sprintf("SELECT COUNT(uuid) FROM %s WHERE uuid = ?", DUES_COLLECTIONS_TABLE))
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer existsStmt.Close()
	currentStmt, err := tx.PrepareContext(ctx, fmt.Sprintf(
		"SELECT action FROM %s WHERE member_uuid = ? AND collection_uuid = ? ORDER BY id DESC LIMIT 1", DUES_LOG_TABLE))
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer currentStmt.Close()
	insertStmt, err := tx.PrepareContext(ctx, fmt.Sprintf(
		"INSERT INTO %s (member_uuid, collection_uuid, action, actor_uuid, created_at, note) VALUES (?, ?, ?, ?, ?, ?)", DUES_LOG_TABLE))
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer insertStmt.Close()

	var encryptedNote interface{}
	if note != "" {
		encryptedNote = common.Encrypt(note)
	}
	now := time.Now().Unix()
	written := 0
	seen := map[string]bool{}
	for _, collectionUUID := range collectionUUIDs {
		if seen[collectionUUID] {
			continue
		}
		seen[collectionUUID] = true

		var count int
		if err := existsStmt.QueryRowContext(ctx, collectionUUID).Scan(&count); err != nil {
			tx.Rollback()
			return 0, err
		}
		if count == 0 {
			tx.Rollback()
			return 0, ErrDuesCollectionNotFound
		}

		current := DuesActionUnpaid
		err := currentStmt.QueryRowContext(ctx, memberUUID, collectionUUID).Scan(&current)
		if err != nil && err != sql.ErrNoRows {
			tx.Rollback()
			return 0, err
		}
		if current == action {
			continue
		}
		if _, err := insertStmt.ExecContext(ctx, memberUUID, collectionUUID, action, actorUUID, now, encryptedNote); err != nil {
			tx.Rollback()
			return 0, err
		}
		written++
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return written, nil
}

// GetDuesLog returns the journal, most recent entry first, with the names of
// the member and of the author of each entry.
func GetDuesLog(ctx context.Context, filter DuesLogFilter) ([]DuesLogEntry, error) {
	ctx, span := tracer.Start(ctx, "GetDuesLog")
	defer span.End()

	query := fmt.Sprintf(
		"SELECT l.id, l.member_uuid, l.collection_uuid, c.season, c.kind, l.action, l.actor_uuid, l.created_at, l.note FROM %s l "+
			"JOIN %s c ON c.uuid = l.collection_uuid",
		DUES_LOG_TABLE, DUES_COLLECTIONS_TABLE)
	conditions := []string{}
	args := []interface{}{}
	if filter.MemberUUID != "" {
		conditions = append(conditions, "l.member_uuid = ?")
		args = append(args, filter.MemberUUID)
	}
	if filter.CollectionUUID != "" {
		conditions = append(conditions, "l.collection_uuid = ?")
		args = append(args, filter.CollectionUUID)
	}
	if filter.ActorUUID != "" {
		conditions = append(conditions, "l.actor_uuid = ?")
		args = append(args, filter.ActorUUID)
	}
	if filter.Season != 0 {
		conditions = append(conditions, "c.season = ?")
		args = append(args, filter.Season)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = DuesLogDefaultLimit
	}
	if limit > DuesLogMaxLimit {
		limit = DuesLogMaxLimit
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	query += " ORDER BY l.id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []DuesLogEntry{}
	uuids := map[string]bool{}
	for rows.Next() {
		var e DuesLogEntry
		var note []byte
		if err := rows.Scan(&e.ID, &e.MemberUUID, &e.CollectionUUID, &e.Season, &e.Kind, &e.Action, &e.ActorUUID, &e.CreatedAt, &note); err != nil {
			return nil, err
		}
		e.Note = decryptOrEmpty(note)
		uuids[e.MemberUUID] = true
		uuids[e.ActorUUID] = true
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	names, err := memberNames(ctx, uuids)
	if err != nil {
		return nil, err
	}
	for i := range entries {
		entries[i].MemberName = names[entries[i].MemberUUID]
		entries[i].ActorName = names[entries[i].ActorUUID]
	}
	return entries, nil
}

// memberNames returns "first last" for each UUID. Unlike Member.Get, it also
// names deleted members, so the journal keeps showing who did what after an
// account is removed. A purged member has no name left and gets none.
func memberNames(ctx context.Context, uuids map[string]bool) (map[string]string, error) {
	names := map[string]string{}
	if len(uuids) == 0 {
		return names, nil
	}
	placeholders := make([]string, 0, len(uuids))
	args := make([]interface{}, 0, len(uuids)+1)
	for uuid := range uuids {
		placeholders = append(placeholders, "?")
		args = append(args, uuid)
	}
	args = append(args, MEMBERSSTATUSPURGED)
	rows, err := db.QueryContext(ctx, fmt.Sprintf(
		"SELECT uuid, firstName, lastName FROM %s WHERE uuid IN (%s) AND status != ?",
		MEMBERSTABLE, strings.Join(placeholders, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var uuid string
		var firstName, lastName []byte
		if err := rows.Scan(&uuid, &firstName, &lastName); err != nil {
			return nil, err
		}
		names[uuid] = strings.TrimSpace(decryptOrEmpty(firstName) + " " + decryptOrEmpty(lastName))
	}
	return names, rows.Err()
}

// decryptOrEmpty decrypts a value that may be missing or unreadable (data of a
// member scrubbed outside of the application).
func decryptOrEmpty(data []byte) (plain string) {
	if len(data) == 0 {
		return ""
	}
	defer func() {
		if r := recover(); r != nil {
			plain = ""
		}
	}()
	return common.Decrypt(data)
}

// GetDuesCollection returns one collection, or ErrDuesCollectionNotFound.
func GetDuesCollection(ctx context.Context, uuid string) (DuesCollection, error) {
	ctx, span := tracer.Start(ctx, "GetDuesCollection")
	defer span.End()

	c := DuesCollection{UUID: uuid}
	err := db.QueryRowContext(ctx, fmt.Sprintf(
		"SELECT season, kind, amount, due_date FROM %s WHERE uuid = ?", DUES_COLLECTIONS_TABLE), uuid).
		Scan(&c.Season, &c.Kind, &c.Amount, &c.DueDate)
	if err == sql.ErrNoRows {
		return c, ErrDuesCollectionNotFound
	}
	return c, err
}

// GetDuesStatus returns the current state of a member for a collection:
// unpaid when nothing was recorded.
func GetDuesStatus(ctx context.Context, memberUUID, collectionUUID string) (string, error) {
	ctx, span := tracer.Start(ctx, "GetDuesStatus")
	defer span.End()

	status := DuesActionUnpaid
	err := db.QueryRowContext(ctx, fmt.Sprintf(
		"SELECT action FROM %s WHERE member_uuid = ? AND collection_uuid = ? ORDER BY id DESC LIMIT 1", DUES_LOG_TABLE),
		memberUUID, collectionUUID).Scan(&status)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	return status, nil
}

// Attendance counts the events at which the presence of a member was
// confirmed. FirstVisit is the start of the first of them (Unix seconds).
type Attendance struct {
	MemberUUID string `json:"memberUuid"`
	FirstVisit int64  `json:"firstVisit"`
	Count      int    `json:"count"`
}

// GetAttendance returns the attendance of one member, or of every member who
// came at least once when memberUUID is empty. Only presences confirmed by an
// admin count: an answer "yes" without confirmation does not.
func GetAttendance(ctx context.Context, memberUUID string) ([]Attendance, error) {
	ctx, span := tracer.Start(ctx, "GetAttendance")
	defer span.End()

	query := fmt.Sprintf(
		"SELECT p.member_uuid, MIN(e.startDate), COUNT(*) FROM %s p "+
			"JOIN %s e ON e.uuid = p.event_uuid "+
			"WHERE p.presence = ? AND e.deleted = 0",
		PARTICIPATION_TABLE, EVENTS_TABLE)
	args := []interface{}{common.AnswerYes}
	if memberUUID != "" {
		query += " AND p.member_uuid = ?"
		args = append(args, memberUUID)
	}
	query += " GROUP BY p.member_uuid"
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attendance := []Attendance{}
	for rows.Next() {
		var a Attendance
		if err := rows.Scan(&a.MemberUUID, &a.FirstVisit, &a.Count); err != nil {
			return nil, err
		}
		attendance = append(attendance, a)
	}
	return attendance, rows.Err()
}

// Keys of the dues settings.
const (
	DuesSettingInteracEmail = "interacEmail"
	DuesSettingCercleURL    = "cercleUrl"
)

// GetDuesSetting returns a setting, or an empty string when it is not set.
func GetDuesSetting(ctx context.Context, key string) (string, error) {
	ctx, span := tracer.Start(ctx, "GetDuesSetting")
	defer span.End()

	var value string
	err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT value FROM %s WHERE key = ?", DUES_SETTINGS_TABLE), key).Scan(&value)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	return value, nil
}

// SetDuesSetting creates or replaces a setting.
func SetDuesSetting(ctx context.Context, key, value string) error {
	ctx, span := tracer.Start(ctx, "SetDuesSetting")
	defer span.End()

	_, err := db.ExecContext(ctx, fmt.Sprintf(
		"INSERT INTO %s (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", DUES_SETTINGS_TABLE),
		key, value)
	return err
}

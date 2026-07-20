package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/vilisseranen/castellers/common"
)

const (
	BADGE_SERIES_TABLE  = "badge_series"
	BADGES_TABLE        = "badges"
	MEMBER_BADGES_TABLE = "member_badges"
)

// Stable badge codes used by automatic awarding. They match the codes seeded
// in the migration and must not change.
const (
	BADGE_CODE_AMUNT = "amunt"
)

// Badge is a single achievement belonging to a series. The human readable
// name and description are not stored: the frontend derives i18n keys from
// Code (e.g. badges.items.<code>.name).
type Badge struct {
	UUID       string `json:"uuid"`
	SeriesUUID string `json:"seriesUuid"`
	Code       string `json:"code"`
	Image      string `json:"image"`
	Position   int    `json:"position"`
}

// BadgeSeries groups badges together (e.g. the welcome series).
type BadgeSeries struct {
	UUID     string  `json:"uuid"`
	Code     string  `json:"code"`
	Position int     `json:"position"`
	Badges   []Badge `json:"badges"`
}

// MemberBadge is a badge unlocked by a member.
type MemberBadge struct {
	MemberUUID string `json:"memberUuid"`
	BadgeUUID  string `json:"badgeUuid"`
	AwardedAt  int64  `json:"awardedAt"`
	AwardedBy  string `json:"awardedBy"`
}

// GetAllBadges returns every series ordered by position, each with its badges.
func GetAllBadges(ctx context.Context) ([]BadgeSeries, error) {
	ctx, span := tracer.Start(ctx, "GetAllBadges")
	defer span.End()

	seriesStmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"SELECT uuid, code, position FROM %s ORDER BY position ASC", BADGE_SERIES_TABLE))
	if err != nil {
		return nil, err
	}
	defer seriesStmt.Close()
	seriesRows, err := seriesStmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer seriesRows.Close()

	series := []BadgeSeries{}
	seriesIndex := map[string]int{}
	for seriesRows.Next() {
		var s BadgeSeries
		if err := seriesRows.Scan(&s.UUID, &s.Code, &s.Position); err != nil {
			return nil, err
		}
		s.Badges = []Badge{}
		seriesIndex[s.UUID] = len(series)
		series = append(series, s)
	}
	if err := seriesRows.Err(); err != nil {
		return nil, err
	}

	badgeStmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"SELECT uuid, series_uuid, code, image, position FROM %s ORDER BY position ASC", BADGES_TABLE))
	if err != nil {
		return nil, err
	}
	defer badgeStmt.Close()
	badgeRows, err := badgeStmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer badgeRows.Close()
	for badgeRows.Next() {
		var b Badge
		if err := badgeRows.Scan(&b.UUID, &b.SeriesUUID, &b.Code, &b.Image, &b.Position); err != nil {
			return nil, err
		}
		if idx, ok := seriesIndex[b.SeriesUUID]; ok {
			series[idx].Badges = append(series[idx].Badges, b)
		}
	}
	if err := badgeRows.Err(); err != nil {
		return nil, err
	}
	return series, nil
}

// Get loads a single badge by UUID. Returns sql.ErrNoRows if not found.
func (b *Badge) Get(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "Badge.Get")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"SELECT series_uuid, code, image, position FROM %s WHERE uuid = ?", BADGES_TABLE))
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.QueryRowContext(ctx, b.UUID).Scan(&b.SeriesUUID, &b.Code, &b.Image, &b.Position)
}

// GetBadgeByCode loads a single badge by its stable code (e.g. "amunt").
// Returns sql.ErrNoRows if no badge has that code.
func GetBadgeByCode(ctx context.Context, code string) (Badge, error) {
	ctx, span := tracer.Start(ctx, "GetBadgeByCode")
	defer span.End()
	b := Badge{Code: code}
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"SELECT uuid, series_uuid, image, position FROM %s WHERE code = ?", BADGES_TABLE))
	if err != nil {
		return b, err
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, code).Scan(&b.UUID, &b.SeriesUUID, &b.Image, &b.Position)
	return b, err
}

// GetMemberBadges returns the badges unlocked by a member.
func GetMemberBadges(ctx context.Context, memberUUID string) ([]MemberBadge, error) {
	ctx, span := tracer.Start(ctx, "GetMemberBadges")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"SELECT badge_uuid, awarded_at, awarded_by FROM %s WHERE member_uuid = ?", MEMBER_BADGES_TABLE))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, memberUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	badges := []MemberBadge{}
	for rows.Next() {
		mb := MemberBadge{MemberUUID: memberUUID}
		var awardedBy sql.NullString
		if err := rows.Scan(&mb.BadgeUUID, &mb.AwardedAt, &awardedBy); err != nil {
			return nil, err
		}
		mb.AwardedBy = nullToEmptyString(awardedBy)
		badges = append(badges, mb)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return badges, nil
}

// GetBadgeMembers returns the UUIDs of the members holding a badge.
func GetBadgeMembers(ctx context.Context, badgeUUID string) ([]string, error) {
	ctx, span := tracer.Start(ctx, "GetBadgeMembers")
	defer span.End()
	stmt, err := db.PrepareContext(ctx, fmt.Sprintf(
		"SELECT member_uuid FROM %s WHERE badge_uuid = ?", MEMBER_BADGES_TABLE))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, badgeUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	memberUUIDs := []string{}
	for rows.Next() {
		var uuid string
		if err := rows.Scan(&uuid); err != nil {
			return nil, err
		}
		memberUUIDs = append(memberUUIDs, uuid)
	}
	return memberUUIDs, rows.Err()
}

// AssignBadge grants a badge to several members. Idempotent: assigning an
// already unlocked badge is a no-op (INSERT OR IGNORE on the unique pair).
// Returns the member UUIDs that were newly assigned (RowsAffected > 0).
func AssignBadge(ctx context.Context, badgeUUID string, memberUUIDs []string, awardedBy string) ([]string, error) {
	ctx, span := tracer.Start(ctx, "AssignBadge")
	defer span.End()
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(
		"INSERT OR IGNORE INTO %s (member_uuid, badge_uuid, awarded_at, awarded_by) VALUES (?, ?, ?, ?)", MEMBER_BADGES_TABLE))
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer stmt.Close()
	now := time.Now().Unix()
	assigned := make([]string, 0, len(memberUUIDs))
	for _, memberUUID := range memberUUIDs {
		result, err := stmt.ExecContext(ctx, memberUUID, badgeUUID, now, stringOrNull(awardedBy))
		if err != nil {
			tx.Rollback()
			common.Error("Error assigning badge %s to %s: %v", badgeUUID, memberUUID, err)
			return nil, err
		}
		n, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if n > 0 {
			assigned = append(assigned, memberUUID)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return assigned, nil
}

// RemoveBadge revokes a badge from several members (fixes mistaken awards).
func RemoveBadge(ctx context.Context, badgeUUID string, memberUUIDs []string) error {
	ctx, span := tracer.Start(ctx, "RemoveBadge")
	defer span.End()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(
		"DELETE FROM %s WHERE badge_uuid = ? AND member_uuid = ?", MEMBER_BADGES_TABLE))
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, memberUUID := range memberUUIDs {
		if _, err := stmt.ExecContext(ctx, badgeUUID, memberUUID); err != nil {
			tx.Rollback()
			common.Error("Error removing badge %s from %s: %v", badgeUUID, memberUUID, err)
			return err
		}
	}
	return tx.Commit()
}

package controller

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

const (
	ERRORGETBADGES       = "error getting badges"
	ERRORGETMEMBERBADGES = "error getting member badges"
	ERRORBADGENOTFOUND   = "badge not found"
	ERRORBADGEMEMBERS    = "memberUuids required"
	ERRORASSIGNBADGE     = "error assigning badge"
	ERRORREMOVEBADGE     = "error removing badge"
)

type badgeMembersRequest struct {
	MemberUUIDs   []string `json:"memberUuids"`
	NotifyByEmail bool     `json:"notifyByEmail"`
}

// GetBadges returns all badge series with their badges (definitions only).
func GetBadges(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetBadges")
	defer span.End()

	series, err := model.GetAllBadges(ctx)
	if err != nil {
		common.Warn("Error getting badges: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETBADGES)
		return
	}
	RespondWithJSON(w, http.StatusOK, series)
}

// GetMemberBadges returns the badges a member has unlocked. Any authenticated
// member can view another member's unlocked badges.
func GetMemberBadges(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetMemberBadges")
	defer span.End()

	vars := mux.Vars(r)
	memberUUID := vars["member_uuid"]

	badges, err := model.GetMemberBadges(ctx, memberUUID)
	if err != nil {
		common.Warn("Error getting member badges: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETMEMBERBADGES)
		return
	}
	RespondWithJSON(w, http.StatusOK, badges)
}

// GetBadgeMembers returns the UUIDs of members holding a badge (admin only).
func GetBadgeMembers(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GetBadgeMembers")
	defer span.End()

	vars := mux.Vars(r)
	badgeUUID := vars["badge_uuid"]

	members, err := model.GetBadgeMembers(ctx, badgeUUID)
	if err != nil {
		common.Warn("Error getting badge members: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORGETMEMBERBADGES)
		return
	}
	RespondWithJSON(w, http.StatusOK, members)
}

// AssignBadge grants a badge to a batch of members (admin only).
func AssignBadge(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "AssignBadge")
	defer span.End()

	vars := mux.Vars(r)
	badgeUUID := vars["badge_uuid"]

	var body badgeMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	if len(body.MemberUUIDs) == 0 {
		RespondWithError(w, http.StatusBadRequest, ERRORBADGEMEMBERS)
		return
	}

	badge := model.Badge{UUID: badgeUUID}
	if err := badge.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, ERRORBADGENOTFOUND)
		default:
			common.Warn("Error getting badge: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORASSIGNBADGE)
		}
		return
	}

	tokenAuth, err := ExtractToken(ctx, r)
	if err != nil {
		common.Warn("Error reading token: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORAUTHENTICATION)
		return
	}

	assigned, err := model.AssignBadge(ctx, badgeUUID, body.MemberUUIDs, tokenAuth.UserId)
	if err != nil {
		common.Warn("Error assigning badge: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORASSIGNBADGE)
		return
	}

	if body.NotifyByEmail && len(assigned) > 0 {
		payload := model.BadgeAwardedPayload{BadgeCode: badge.Code, MemberUUIDs: assigned}
		payloadBytes := new(bytes.Buffer)
		if err := json.NewEncoder(payloadBytes).Encode(payload); err != nil {
			common.Warn("Error encoding badge awarded notification: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
			return
		}
		n := model.Notification{
			NotificationType: model.TypeBadgeAwarded,
			ObjectUUID:       badgeUUID,
			SendDate:         int(time.Now().Unix()),
			Payload:          payloadBytes.Bytes(),
		}
		if err := n.CreateNotification(ctx); err != nil {
			common.Warn("Error creating badge awarded notification: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORNOTIFICATION)
			return
		}
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// awardBadgeByCode grants a badge (identified by its stable code) to a single
// member as an automatic, self-earned award (awarded_by left empty).
//
// It is best-effort: any failure is logged but never returned, so an automatic
// award can never break the action that triggered it. Assignment is idempotent,
// so calling it repeatedly is harmless.
func awardBadgeByCode(ctx context.Context, code, memberUUID string) {
	badge, err := model.GetBadgeByCode(ctx, code)
	if err != nil {
		if err != sql.ErrNoRows {
			common.Warn("Error looking up badge %q for auto-award: %s", code, err.Error())
		}
		return
	}
	if _, err := model.AssignBadge(ctx, badge.UUID, []string{memberUUID}, ""); err != nil {
		common.Warn("Error auto-awarding badge %q to %s: %s", code, memberUUID, err.Error())
	}
}

// RemoveBadge revokes a badge from a batch of members (admin only).
func RemoveBadge(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RemoveBadge")
	defer span.End()

	vars := mux.Vars(r)
	badgeUUID := vars["badge_uuid"]

	var body badgeMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondWithError(w, http.StatusBadRequest, ERRORINVALIDPAYLOAD)
		return
	}
	if len(body.MemberUUIDs) == 0 {
		RespondWithError(w, http.StatusBadRequest, ERRORBADGEMEMBERS)
		return
	}

	badge := model.Badge{UUID: badgeUUID}
	if err := badge.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, ERRORBADGENOTFOUND)
		default:
			common.Warn("Error getting badge: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORREMOVEBADGE)
		}
		return
	}

	if err := model.RemoveBadge(ctx, badgeUUID, body.MemberUUIDs); err != nil {
		common.Warn("Error removing badge: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORREMOVEBADGE)
		return
	}
	RespondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

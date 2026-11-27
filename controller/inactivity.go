package controller

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/mail"
	"github.com/vilisseranen/castellers/model"
)

const ERRORKEEPMEMBER = "error keeping member"

const day = int64(24 * 3600)

// RunInactiveMembersScanOnce runs the scan of the inactive members once (used by tests and cron).
func RunInactiveMembersScanOnce() {
	deleteInactiveMembers()
}

// neverActivated tells an account that was created and never used: the member
// did not choose a password and never answered nor came to an event.
func neverActivated(a model.MemberActivity) bool {
	return a.Status == model.MEMBERSSTATUSCREATED && a.LastParticipation == 0
}

// DeletionDue tells if a member has been inactive for too long to be kept:
//   - an account never activated, pending_delete_delay_days after its creation;
//   - a member in pause, or with an account that was not activated but who came
//     anyway, inactive_delete_delay_days after their last sign of life (answer or
//     presence to an event, reactivation or "keep" by an admin, creation).
//
// Active members and admins are never due.
func DeletionDue(a model.MemberActivity, now int64) bool {
	if a.Type == model.MEMBERSTYPEADMIN {
		return false
	}
	if a.Status != model.MEMBERSSTATUSCREATED && a.Status != model.MEMBERSSTATUSPAUSED {
		return false
	}
	lastSeen := a.CreatedAt
	if a.LastActivityDate > lastSeen {
		lastSeen = a.LastActivityDate
	}
	if a.LastParticipation > lastSeen {
		lastSeen = a.LastParticipation
	}
	delay := int64(common.GetConfigInt("inactive_delete_delay_days")) * day
	if neverActivated(a) {
		delay = int64(common.GetConfigInt("pending_delete_delay_days")) * day
	}
	return now-lastSeen >= delay
}

// deleteInactiveMembers looks for the members who have been inactive for too
// long. They are first told that their account will be deleted, then deleted
// deletion_notice_days later if nothing happened in between. A deleted member
// can still be restored by an admin until they are purged.
func deleteInactiveMembers() {
	ctx, span := tracer.Start(context.Background(), "deleteInactiveMembers")
	defer span.End()

	if !common.GetConfigBool("auto_delete_enabled") {
		return
	}
	activities, err := model.GetMembersActivity(ctx)
	if err != nil {
		common.Error("Cannot get the activity of the members: %v", err)
		return
	}
	now := time.Now().Unix()
	for _, activity := range activities {
		member := model.Member{UUID: activity.UUID}
		if !DeletionDue(activity, now) {
			// They came back, or somebody kept them
			if activity.DeletionNoticeAt > 0 {
				if err := member.ClearDeletionNotice(ctx); err != nil {
					common.Error("Cannot cancel the deletion of member %s: %v", activity.UUID, err)
				}
			}
			continue
		}
		if activity.DeletionNoticeAt == 0 {
			if err := member.SetDeletionNotice(ctx, now); err != nil {
				common.Error("Cannot plan the deletion of member %s: %v", activity.UUID, err)
				continue
			}
			common.Info("Member %s is inactive: deletion planned", activity.UUID)
			sendDeletionNotice(ctx, activity, now+model.DeletionNoticeDelay())
			continue
		}
		if now-activity.DeletionNoticeAt >= model.DeletionNoticeDelay() {
			if err := member.DeleteMember(ctx); err != nil {
				common.Error("Cannot delete inactive member %s: %v", activity.UUID, err)
				continue
			}
			if err := revokeMemberSessions(ctx, activity.UUID, ""); err != nil {
				common.Error("Cannot revoke the sessions of member %s: %s", activity.UUID, err.Error())
			}
			common.Info("Inactive member %s was deleted", activity.UUID)
		}
	}
}

// sendDeletionNotice tells a member that their account will be deleted. A
// member without an email (canalla) is announced to the members responsible
// for them. Nobody can be told for a guest: the admins will be.
func sendDeletionNotice(ctx context.Context, activity model.MemberActivity, deletionAt int64) {
	if !common.GetConfigBool("smtp.enabled") {
		return
	}
	member := model.Member{UUID: activity.UUID}
	if err := member.Get(ctx); err != nil {
		common.Error("Cannot get member %s to tell them of their deletion: %v", activity.UUID, err)
		return
	}
	recipients := []model.Member{member}
	if member.Email == "" {
		responsibles, err := member.GetResponsibles(ctx)
		if err != nil {
			common.Error("Cannot get who is responsible for member %s: %v", activity.UUID, err)
			return
		}
		recipients = responsibles
	}
	for _, recipient := range recipients {
		if recipient.Email == "" {
			continue
		}
		payload := mail.EmailDeletionNoticePayload{
			Recipient:      recipient,
			Member:         member,
			DeletionAt:     deletionAt,
			NeverActivated: neverActivated(activity),
		}
		if err := mail.SendDeletionNoticeEmail(ctx, payload); err != nil {
			common.Error("Cannot tell member %s of the deletion of member %s: %v", recipient.UUID, member.UUID, err)
		}
	}
}

// sendDeletionDigest tells the admins, in one email, about the members whose
// deletion was planned and about the deleted members whose data will be erased
// within deletion_notice_days. Each member is announced once. Nothing is sent
// when there is nothing new.
func sendDeletionDigest() {
	ctx, span := tracer.Start(context.Background(), "sendDeletionDigest")
	defer span.End()

	if !common.GetConfigBool("smtp.enabled") {
		return
	}
	deletions, purges, err := model.GetAdminNotices(ctx, time.Now().Unix()+model.DeletionNoticeDelay())
	if err != nil {
		common.Error("Cannot get what to tell the admins: %v", err)
		return
	}
	if len(deletions) == 0 && len(purges) == 0 {
		return
	}
	m := model.Member{}
	admins, err := m.GetAll(ctx, []string{}, []string{model.MEMBERSTYPEADMIN})
	if err != nil {
		common.Error("Cannot get the admins: %v", err)
		return
	}
	sent := 0
	for _, admin := range admins {
		if admin.Subscribed != 1 || admin.Email == "" {
			continue
		}
		payload := mail.EmailDeletionDigestPayload{Member: admin, Deletions: deletions, Purges: purges}
		if err := mail.SendDeletionDigestEmail(ctx, payload); err != nil {
			common.Error("Cannot send the digest of deletions to admin %s: %v", admin.UUID, err)
			continue
		}
		sent++
	}
	if sent == 0 {
		// Nobody was told: try again next time
		return
	}
	uuids := []string{}
	for _, notice := range append(deletions, purges...) {
		uuids = append(uuids, notice.UUID)
	}
	if err := model.SetAdminNotified(ctx, uuids); err != nil {
		common.Error("Cannot note that the admins were told: %v", err)
	}
}

// KeepMember lets an admin keep a member who would be deleted for inactivity:
// their inactivity delay starts again. Their status does not change.
func KeepMember(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "KeepMember")
	defer span.End()

	vars := mux.Vars(r)
	UUID := vars["member_uuid"]
	m := model.Member{UUID: UUID}
	if err := m.Get(ctx); err != nil {
		switch err {
		case sql.ErrNoRows:
			RespondWithError(w, http.StatusNotFound, ERRORMEMBERNOTFOUND)
		default:
			common.Warn("Error getting member: %s", err.Error())
			RespondWithError(w, http.StatusInternalServerError, ERRORKEEPMEMBER)
		}
		return
	}
	if err := m.KeepMember(ctx); err != nil {
		common.Warn("Error keeping member: %s", err.Error())
		RespondWithError(w, http.StatusInternalServerError, ERRORKEEPMEMBER)
		return
	}
	RespondWithJSON(w, http.StatusOK, m)
}

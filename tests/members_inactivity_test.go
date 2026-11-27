package tests

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/vilisseranen/castellers/controller"
	"github.com/vilisseranen/castellers/model"
)

const testDay = int64(24 * 3600)

func (test *TestHelper) execSQL(query string, args ...interface{}) {
	db, err := sql.Open("sqlite3", testDbName)
	if err != nil {
		tFatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(query, args...); err != nil {
		tFatal(err)
	}
}

// deletionNoticeAt returns when the member was told of their deletion, 0 if not.
func (test *TestHelper) deletionNoticeAt(uuid string) int64 {
	return int64(test.countRows("SELECT COALESCE(deletion_notice_at, 0) FROM members WHERE uuid = ?", uuid))
}

func TestDeletionDue(t *testing.T) {
	now := time.Now().Unix()
	cases := []struct {
		name     string
		activity model.MemberActivity
		due      bool
	}{
		{"account never activated, created 181 days ago",
			model.MemberActivity{Type: model.MEMBERSTYPEREGULAR, Status: model.MEMBERSSTATUSCREATED, CreatedAt: now - 181*testDay}, true},
		{"account never activated, created 179 days ago",
			model.MemberActivity{Type: model.MEMBERSTYPEREGULAR, Status: model.MEMBERSSTATUSCREATED, CreatedAt: now - 179*testDay}, false},
		{"account never activated, creation unknown",
			model.MemberActivity{Type: model.MEMBERSTYPEREGULAR, Status: model.MEMBERSSTATUSCREATED}, true},
		{"account never activated, kept by an admin last month",
			model.MemberActivity{Type: model.MEMBERSTYPEREGULAR, Status: model.MEMBERSSTATUSCREATED, CreatedAt: now - 400*testDay, LastActivityDate: now - 30*testDay}, false},
		{"account not activated, but came 200 days ago",
			model.MemberActivity{Type: model.MEMBERSTYPEREGULAR, Status: model.MEMBERSSTATUSCREATED, CreatedAt: now - 400*testDay, LastParticipation: now - 200*testDay}, false},
		{"account not activated, came 731 days ago",
			model.MemberActivity{Type: model.MEMBERSTYPEREGULAR, Status: model.MEMBERSSTATUSCREATED, CreatedAt: now - 900*testDay, LastParticipation: now - 731*testDay}, true},
		{"in pause, last seen 731 days ago",
			model.MemberActivity{Type: model.MEMBERSTYPEREGULAR, Status: model.MEMBERSSTATUSPAUSED, CreatedAt: now - 2000*testDay, LastParticipation: now - 731*testDay}, true},
		{"in pause, last seen 729 days ago",
			model.MemberActivity{Type: model.MEMBERSTYPEREGULAR, Status: model.MEMBERSSTATUSPAUSED, CreatedAt: now - 2000*testDay, LastParticipation: now - 729*testDay}, false},
		{"in pause, reactivated by an admin last year",
			model.MemberActivity{Type: model.MEMBERSTYPEREGULAR, Status: model.MEMBERSSTATUSPAUSED, LastParticipation: now - 1500*testDay, LastActivityDate: now - 365*testDay}, false},
		{"in pause, created 200 days ago and never came",
			model.MemberActivity{Type: model.MEMBERSTYPEREGULAR, Status: model.MEMBERSSTATUSPAUSED, CreatedAt: now - 200*testDay}, false},
		{"guest in pause, last seen 3 years ago",
			model.MemberActivity{Type: model.MEMBERSTYPEGUEST, Status: model.MEMBERSSTATUSPAUSED, LastParticipation: now - 1095*testDay}, true},
		{"canalla in pause, last seen 3 years ago",
			model.MemberActivity{Type: model.MEMBERSTYPECANALLA, Status: model.MEMBERSSTATUSPAUSED, LastParticipation: now - 1095*testDay}, true},
		{"active member, nothing known",
			model.MemberActivity{Type: model.MEMBERSTYPEREGULAR, Status: model.MEMBERSSTATUSACTIVATED}, false},
		{"admin in pause, last seen 3 years ago",
			model.MemberActivity{Type: model.MEMBERSTYPEADMIN, Status: model.MEMBERSSTATUSPAUSED, LastParticipation: now - 1095*testDay}, false},
		{"admin who never activated their account",
			model.MemberActivity{Type: model.MEMBERSTYPEADMIN, Status: model.MEMBERSSTATUSCREATED}, false},
	}
	for _, c := range cases {
		if due := controller.DeletionDue(c.activity, now); due != c.due {
			t.Errorf("%s: expected due = %v. Got %v", c.name, c.due, due)
		}
	}
}

// An inactive member is told first, and deleted 30 days later.
func TestInactiveMemberIsToldThenDeleted(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	now := time.Now().Unix()
	// In pause, not seen for more than 2 years
	h.addMember("abc123", "Joan", "Marti", "", "", "", "", "member", "joan@marti.ca", "")
	h.execSQL("UPDATE members SET status = 'paused', created_at = ? WHERE uuid = 'abc123'", now-800*testDay)
	// Account created 7 months ago, never activated
	h.addMember("abc456", "Nuria", "Vila", "", "", "", "", "member", "nuria@vila.ca", "")
	h.execSQL("UPDATE members SET created_at = ? WHERE uuid = 'abc456'", now-210*testDay)
	// Account created last month
	h.addMember("abc789", "Pau", "Roca", "", "", "", "", "member", "pau@roca.ca", "")
	h.execSQL("UPDATE members SET created_at = ? WHERE uuid = 'abc789'", now-30*testDay)
	// In pause for a year
	h.addMember("abcabc", "Laia", "Font", "", "", "", "", "member", "laia@font.ca", "")
	h.execSQL("UPDATE members SET status = 'paused', created_at = ? WHERE uuid = 'abcabc'", now-365*testDay)

	controller.RunInactiveMembersScanOnce()

	for _, uuid := range []string{"abc123", "abc456"} {
		if notice := h.deletionNoticeAt(uuid); notice < now || notice > time.Now().Unix() {
			t.Errorf("Expected member %s to be told now. Got %d", uuid, notice)
		}
	}
	for _, uuid := range []string{"abc789", "abcabc", "deadfeed"} {
		if notice := h.deletionNoticeAt(uuid); notice != 0 {
			t.Errorf("Expected no planned deletion for member %s. Got %d", uuid, notice)
		}
	}
	if status := h.getMemberStatus("abc123"); status != model.MEMBERSSTATUSPAUSED {
		t.Errorf("Expected the member to stay in pause during the notice. Got '%s'", status)
	}
	// The list of members gives the day of the deletion
	members := h.listMembers(t, adminToken, "?status=paused,created")
	notice := h.deletionNoticeAt("abc123")
	if deletionAt := int64(members["abc123"]["deletionAt"].(float64)); deletionAt != notice+30*testDay {
		t.Errorf("Expected the deletion 30 days after the notice. Got %d seconds after", deletionAt-notice)
	}
	if members["abc789"]["deletionAt"] != float64(0) {
		t.Errorf("Expected no deletion planned for a recent account. Got %v", members["abc789"]["deletionAt"])
	}

	// Scanning again does not move the notice
	h.execSQL("UPDATE members SET deletion_notice_at = ? WHERE uuid IN ('abc123', 'abc456')", now-10*testDay)
	controller.RunInactiveMembersScanOnce()
	if notice := h.deletionNoticeAt("abc123"); notice != now-10*testDay {
		t.Errorf("Expected the notice to stay 10 days ago. Got %d", notice)
	}
	if status := h.getMemberStatus("abc123"); status != model.MEMBERSSTATUSPAUSED {
		t.Errorf("Expected the member to stay in pause 10 days after the notice. Got '%s'", status)
	}

	// 30 days after the notice: deleted, and to be purged 90 days later
	h.execSQL("UPDATE members SET deletion_notice_at = ? WHERE uuid IN ('abc123', 'abc456')", now-30*testDay-60)
	controller.RunInactiveMembersScanOnce()
	for _, uuid := range []string{"abc123", "abc456"} {
		if status := h.getMemberStatus(uuid); status != model.MEMBERSSTATUSDELETED {
			t.Errorf("Expected member %s to be deleted. Got '%s'", uuid, status)
		}
		if h.countRows("SELECT COUNT(*) FROM members WHERE uuid = ? AND deleted_at >= ? AND deletion_notice_at IS NULL", uuid, now) != 1 {
			t.Errorf("Expected member %s to be deleted now, without a notice left", uuid)
		}
	}
	for _, uuid := range []string{"abc789", "abcabc"} {
		if status := h.getMemberStatus(uuid); status == model.MEMBERSSTATUSDELETED {
			t.Errorf("Expected member %s not to be deleted", uuid)
		}
	}
	deleted := h.listMembers(t, adminToken, "?status=deleted")
	if _, listed := deleted["abc123"]; !listed || len(deleted) != 2 {
		t.Errorf("Expected the 2 deleted members to be listed, to restore them. Got %d", len(deleted))
	}

	// Restored by an admin: not told again right away
	if code, _ := h.post("/members/abc123/restore", adminToken, ""); code != http.StatusOK {
		t.Fatalf("Expected to restore the member. Got %d", code)
	}
	controller.RunInactiveMembersScanOnce()
	if notice := h.deletionNoticeAt("abc123"); notice != 0 {
		t.Errorf("Expected no planned deletion for a member just restored. Got %d", notice)
	}
	if status := h.getMemberStatus("abc123"); status == model.MEMBERSSTATUSDELETED {
		t.Error("Expected the restored member not to be deleted again")
	}
}

// An admin keeps a member who was going to be deleted.
func TestKeepMember(t *testing.T) {
	h.clearTables()
	adminToken := h.addAnAdmin()
	memberToken := h.addAMember()
	now := time.Now().Unix()
	h.addMember("abc123", "Joan", "Marti", "", "", "", "", "guest", "", "")
	h.execSQL("UPDATE members SET status = 'paused', created_at = ? WHERE uuid = 'abc123'", now-800*testDay)
	controller.RunInactiveMembersScanOnce()
	if notice := h.deletionNoticeAt("abc123"); notice == 0 {
		t.Fatal("Expected the deletion of the guest to be planned")
	}

	if code, _ := h.post("/members/abc123/keep", memberToken, ""); code != http.StatusUnauthorized {
		t.Errorf("Expected 401 when a member keeps somebody. Got %d", code)
	}
	if code, _ := h.post("/members/ffff/keep", adminToken, ""); code != http.StatusNotFound {
		t.Errorf("Expected 404 when keeping an unknown member. Got %d", code)
	}
	if code, _ := h.post("/members/abc123/keep", adminToken, ""); code != http.StatusOK {
		t.Fatalf("Expected to keep the member. Got %d", code)
	}
	if notice := h.deletionNoticeAt("abc123"); notice != 0 {
		t.Errorf("Expected the deletion to be cancelled. Got %d", notice)
	}
	if status := h.getMemberStatus("abc123"); status != model.MEMBERSSTATUSPAUSED {
		t.Errorf("Expected the status not to change. Got '%s'", status)
	}
	// The delay starts again
	controller.RunInactiveMembersScanOnce()
	if notice := h.deletionNoticeAt("abc123"); notice != 0 {
		t.Errorf("Expected no planned deletion after the member was kept. Got %d", notice)
	}
}

// A member who logs in, or answers to an event, after the notice stays.
func TestSignOfLifeCancelsDeletion(t *testing.T) {
	h.clearTables()
	now := time.Now().Unix()
	h.addAMember()
	h.addMember("abc123", "Joan", "Marti", "", "", "", "", "member", "joan@marti.ca", "")
	h.execSQL("UPDATE members SET status = 'paused', created_at = ? WHERE uuid IN ('deadbeef', 'abc123')", now-800*testDay)
	controller.RunInactiveMembersScanOnce()
	for _, uuid := range []string{"deadbeef", "abc123"} {
		if notice := h.deletionNoticeAt(uuid); notice == 0 {
			t.Fatalf("Expected the deletion of member %s to be planned", uuid)
		}
	}

	// Logging in cancels the deletion at once
	if token := h.login("member", "member"); token == "" {
		t.Fatal("Expected the member to log in")
	}
	if notice := h.deletionNoticeAt("deadbeef"); notice != 0 {
		t.Errorf("Expected the deletion to be cancelled by the login. Got %d", notice)
	}
	if h.countRows("SELECT COUNT(*) FROM members WHERE uuid = 'deadbeef' AND last_activity_date >= ?", now) != 1 {
		t.Error("Expected the login to count as an activity")
	}

	// An answer to an event, even "no", cancels it at the next scan
	h.addEvent("e1", "Practice", int(now+7*testDay), int(now+7*testDay+3600))
	h.addParticipation("abc123", "e1", "no")
	controller.RunInactiveMembersScanOnce()
	for _, uuid := range []string{"deadbeef", "abc123"} {
		if notice := h.deletionNoticeAt(uuid); notice != 0 {
			t.Errorf("Expected no planned deletion for member %s. Got %d", uuid, notice)
		}
		if status := h.getMemberStatus(uuid); status == model.MEMBERSSTATUSDELETED {
			t.Errorf("Expected member %s not to be deleted", uuid)
		}
	}

	// A login without a notice changes nothing
	h.execSQL("UPDATE members SET last_activity_date = 5 WHERE uuid = 'deadbeef'")
	h.login("member", "member")
	if h.countRows("SELECT last_activity_date FROM members WHERE uuid = 'deadbeef'") != 5 {
		t.Error("Expected a login without a notice not to touch the activity")
	}
}

// The admins are told once about each planned deletion and each coming purge.
func TestAdminNotices(t *testing.T) {
	h.clearTables()
	h.addAnAdmin()
	now := time.Now().Unix()
	h.addMember("abc123", "Joan", "Marti", "", "", "", "", "member", "joan@marti.ca", "")
	h.execSQL("UPDATE members SET status = 'paused', created_at = ? WHERE uuid = 'abc123'", now-800*testDay)
	// Deleted 70 days ago: purged in 20 days
	h.addMember("abc456", "Nuria", "Vila", "", "", "", "", "member", "nuria@vila.ca", "")
	h.setMemberDeletedAt("abc456", now-70*testDay)
	// Deleted yesterday: purged in 89 days, too far to tell
	h.addMember("abc789", "Pau", "Roca", "", "", "", "", "member", "pau@roca.ca", "")
	h.setMemberDeletedAt("abc789", now-testDay)
	controller.RunInactiveMembersScanOnce()

	deletions, purges, err := model.GetAdminNotices(t.Context(), now+30*testDay)
	if err != nil {
		t.Fatal(err)
	}
	if len(deletions) != 1 || deletions[0].UUID != "abc123" || deletions[0].Name != "Joan Marti" {
		t.Fatalf("Expected the planned deletion of Joan Marti. Got %v", deletions)
	}
	if deletions[0].Date != h.deletionNoticeAt("abc123")+30*testDay {
		t.Errorf("Expected the day of the deletion. Got %d", deletions[0].Date)
	}
	if len(purges) != 1 || purges[0].UUID != "abc456" || purges[0].Name != "Nuria Vila" || purges[0].Date != now+20*testDay {
		t.Fatalf("Expected the coming purge of Nuria Vila in 20 days. Got %v", purges)
	}

	// Once told, not told again
	if err := model.SetAdminNotified(t.Context(), []string{"abc123", "abc456"}); err != nil {
		t.Fatal(err)
	}
	deletions, purges, err = model.GetAdminNotices(t.Context(), now+30*testDay)
	if err != nil || len(deletions) != 0 || len(purges) != 0 {
		t.Errorf("Expected nothing new to tell. Got %v, %v, %v", deletions, purges, err)
	}
	controller.RunInactiveMembersScanOnce()
	deletions, _, _ = model.GetAdminNotices(t.Context(), now+30*testDay)
	if len(deletions) != 0 {
		t.Errorf("Expected a scan not to announce the same deletion again. Got %v", deletions)
	}

	// When the member is deleted for good, their purge is announced in turn
	h.execSQL("UPDATE members SET deletion_notice_at = ? WHERE uuid = 'abc123'", now-31*testDay)
	controller.RunInactiveMembersScanOnce()
	_, purges, _ = model.GetAdminNotices(t.Context(), now+30*testDay)
	if len(purges) != 0 {
		t.Errorf("Expected nothing to tell right after the deletion. Got %v", purges)
	}
	_, purges, _ = model.GetAdminNotices(t.Context(), now+91*testDay)
	if len(purges) != 2 || purges[0].UUID != "abc123" {
		t.Errorf("Expected the purge of the member just deleted to be announced when it comes, with the one of Pau. Got %v", purges)
	}
}

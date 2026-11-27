package mail

import (
	"strings"
	"testing"
	"time"

	"github.com/vilisseranen/castellers/model"
)

func deletionTestDay(t *testing.T) int64 {
	return time.Date(2026, time.December, 27, 12, 0, 0, 0, montreal(t)).Unix()
}

func TestDeletionNoticeEmail(t *testing.T) {
	member := model.Member{UUID: "m1", FirstName: "Ramon", Email: "ramon@example.com", Language: "fr"}
	payload := EmailDeletionNoticePayload{Recipient: member, Member: member, DeletionAt: deletionTestDay(t)}

	email := deletionNoticeEmail(payload, 24, 90)
	if email.Header.Title != "Ton compte sera supprimé le 27-12-2026" {
		t.Errorf("Unexpected subject: %q", email.Header.Title)
	}
	if email.Top.To != "ramon@example.com" || !strings.Contains(email.Top.Title, "Ramon") {
		t.Errorf("Expected the email to go to Ramon. Got %q, %q", email.Top.To, email.Top.Title)
	}
	text := email.MainSections[0].Text
	for _, expected := range []string{"depuis plus de 24 mois", "90 jours plus tard"} {
		if !strings.Contains(text, expected) {
			t.Errorf("Expected the email to contain %q. Got %q", expected, text)
		}
	}
	if len(email.Actions) != 1 || len(email.Actions[0].Buttons) != 1 || !strings.HasSuffix(email.Actions[0].Buttons[0].Link, "/login") {
		t.Errorf("Expected a button to log in. Got %v", email.Actions)
	}

	// The same email is written in every language, and fits in the template
	for _, language := range []string{"fr", "en", "cat"} {
		payload.Recipient.Language = language
		email := deletionNoticeEmail(payload, 24, 90)
		if strings.Contains(email.Header.Title+email.MainSections[0].Text, "%!") {
			t.Errorf("Badly formatted text in %s: %q / %q", language, email.Header.Title, email.MainSections[0].Text)
		}
		if _, err := email.buildEmail(); err != nil {
			t.Errorf("Cannot build the email in %s: %v", language, err)
		}
	}
	payload.Recipient.Language = "en"
	if title := deletionNoticeEmail(payload, 24, 90).Header.Title; title != "Your account will be deleted on 27-12-2026" {
		t.Errorf("Unexpected subject in English: %q", title)
	}
}

// The parent of a canalla is told about the account of their child.
func TestDeletionNoticeEmailForDependent(t *testing.T) {
	parent := model.Member{UUID: "m1", FirstName: "Ramon", Email: "ramon@example.com", Language: "fr"}
	child := model.Member{UUID: "m2", FirstName: "Pol <b>", LastName: "Petit"}
	payload := EmailDeletionNoticePayload{Recipient: parent, Member: child, DeletionAt: deletionTestDay(t)}

	email := deletionNoticeEmail(payload, 24, 90)
	if email.Top.To != "ramon@example.com" {
		t.Errorf("Expected the email to go to the parent. Got %q", email.Top.To)
	}
	if email.Header.Title != "Le compte de Pol <b> sera supprimé le 27-12-2026" {
		t.Errorf("Unexpected subject: %q", email.Header.Title)
	}
	text := email.MainSections[0].Text
	if !strings.Contains(text, "Nous n&#39;avons pas vu Pol &lt;b&gt; depuis plus de 24 mois") {
		t.Errorf("Expected the email to name the child, escaped. Got %q", text)
	}
	if len(email.Actions) != 0 {
		t.Errorf("Expected no button: a canalla cannot log in. Got %v", email.Actions)
	}
	for _, language := range []string{"fr", "en", "cat"} {
		payload.Recipient.Language = language
		email := deletionNoticeEmail(payload, 24, 90)
		if strings.Contains(email.Header.Title+email.MainSections[0].Text, "%!") {
			t.Errorf("Badly formatted text in %s: %q / %q", language, email.Header.Title, email.MainSections[0].Text)
		}
		if !strings.Contains(email.MainSections[0].Text, "24") {
			t.Errorf("Expected the delay in the text in %s. Got %q", language, email.MainSections[0].Text)
		}
		if _, err := email.buildEmail(); err != nil {
			t.Errorf("Cannot build the email in %s: %v", language, err)
		}
	}
}

func TestDeletionDigestEmail(t *testing.T) {
	admin := model.Member{UUID: "a1", FirstName: "Romà", Email: "roma@example.com", Language: "fr"}
	day := deletionTestDay(t)
	payload := EmailDeletionDigestPayload{
		Member:    admin,
		Deletions: []model.AdminNotice{{UUID: "m1", Name: "Joan Marti", Date: day}, {UUID: "m2", Name: "Nil <i>", Date: day}},
		Purges:    []model.AdminNotice{{UUID: "m3", Name: "Nuria Vila", Date: day}},
	}

	email := deletionDigestEmail(payload)
	if len(email.MainSections) != 2 {
		t.Fatalf("Expected a section for the deletions and one for the purges. Got %d", len(email.MainSections))
	}
	deletions := email.MainSections[0].Text
	for _, expected := range []string{"<li>Joan Marti — 27-12-2026</li>", "<li>Nil &lt;i&gt; — 27-12-2026</li>"} {
		if !strings.Contains(deletions, expected) {
			t.Errorf("Expected the deletions to contain %q. Got %q", expected, deletions)
		}
	}
	if strings.Contains(deletions, "Nuria") || !strings.Contains(email.MainSections[1].Text, "<li>Nuria Vila — 27-12-2026</li>") {
		t.Errorf("Expected Nuria in the purges only. Got %q / %q", deletions, email.MainSections[1].Text)
	}
	if !strings.HasSuffix(email.Actions[0].Buttons[0].Link, "/members?upcoming=1") {
		t.Errorf("Expected a button to the list of members. Got %q", email.Actions[0].Buttons[0].Link)
	}

	// Only what there is to tell
	payload.Deletions = nil
	email = deletionDigestEmail(payload)
	if len(email.MainSections) != 1 || !strings.Contains(email.MainSections[0].Text, "Nuria Vila") {
		t.Errorf("Expected only the purges. Got %v", email.MainSections)
	}
	for _, language := range []string{"fr", "en", "cat"} {
		payload.Member.Language = language
		if _, err := deletionDigestEmail(payload).buildEmail(); err != nil {
			t.Errorf("Cannot build the email in %s: %v", language, err)
		}
	}
}

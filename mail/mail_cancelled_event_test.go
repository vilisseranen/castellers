package mail

import (
	"strings"
	"testing"
	"time"

	"github.com/vilisseranen/castellers/model"
)

func cancelledPayload(event model.Event, language string, restored bool) EmailCancelledEventPayload {
	return EmailCancelledEventPayload{
		Member:   model.Member{UUID: "m1", FirstName: "Ramon", Email: "ramon@example.com", Language: language},
		Event:    event,
		Restored: restored,
	}
}

func TestCancelledEventEmail(t *testing.T) {
	location := montreal(t)
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, location)
	event := practice(location)
	event.Cancelled = 1
	event.CancelReason = model.EVENTCANCELREASONWEATHER
	event.CancelNote = "Orage prévu <à 18 h>\nOn se reprend jeudi."

	email := cancelledEventEmail(cancelledPayload(event, "fr", false), location, now)
	if email.Header.Title != "Annulé : Entraînement" {
		t.Errorf("Unexpected subject: %q", email.Header.Title)
	}
	text := email.MainSections[0].Text
	for _, expected := range []string{
		formatWhen(event, location, now, "fr"),
		"Cet évènement est annulé.",
		"<strong>Raison :</strong> Météo",
		"Orage prévu &lt;à 18 h&gt;<br/>On se reprend jeudi.",
	} {
		if !strings.Contains(text, expected) {
			t.Errorf("Expected the email to contain %q. Got %q", expected, text)
		}
	}
	if _, err := email.buildEmail(); err != nil {
		t.Errorf("Cannot build the email: %v", err)
	}

	// Without a note, only the reason is given.
	event.CancelNote = ""
	event.CancelReason = model.EVENTCANCELREASONTECHNICAL
	text = cancelledEventEmail(cancelledPayload(event, "en", false), location, now).MainSections[0].Text
	if !strings.HasSuffix(text, "<strong>Reason:</strong> Technical difficulties") {
		t.Errorf("Unexpected text without a note: %q", text)
	}
}

func TestRestoredEventEmail(t *testing.T) {
	location := montreal(t)
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, location)
	event := practice(location)

	email := cancelledEventEmail(cancelledPayload(event, "fr", true), location, now)
	if email.Header.Title != "Maintenu : Entraînement" {
		t.Errorf("Unexpected subject: %q", email.Header.Title)
	}
	text := email.MainSections[0].Text
	if !strings.Contains(text, "il a finalement lieu") || strings.Contains(text, "Raison") {
		t.Errorf("Unexpected text: %q", text)
	}
	if len(email.Actions) != 1 || len(email.Actions[0].Buttons) != 1 ||
		!strings.HasSuffix(email.Actions[0].Buttons[0].Link, "/eventShow/e1") {
		t.Errorf("Expected a link to the event. Got %+v", email.Actions)
	}
}

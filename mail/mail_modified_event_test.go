package mail

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

// The translations and the email template are read from the root of the
// repository.
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	common.InitializeTranslations()
	os.Exit(m.Run())
}

func montreal(t *testing.T) *time.Location {
	t.Helper()
	location, err := time.LoadLocation("America/Montreal")
	if err != nil {
		t.Fatal(err)
	}
	return location
}

// practice is a practice on Thursday, November 19, 2026, from 7 PM to 9 PM.
func practice(location *time.Location) model.Event {
	return model.Event{
		UUID:         "e1",
		Name:         "Entraînement",
		Type:         "practice",
		Description:  "Ancienne description.",
		StartDate:    uint(time.Date(2026, time.November, 19, 19, 0, 0, 0, location).Unix()),
		EndDate:      uint(time.Date(2026, time.November, 19, 21, 0, 0, 0, location).Unix()),
		LocationName: "Casal",
		Location:     model.LatLng{Lat: 45.5, Lng: -73.6},
	}
}

func oneDayLater(event model.Event) model.Event {
	event.StartDate += 24 * 3600
	event.EndDate += 24 * 3600
	return event
}

func modifiedPayload(before, after model.Event, language string) EmailModifiedPayload {
	return EmailModifiedPayload{
		Member:            model.Member{UUID: "m1", FirstName: "Ramon", Email: "ramon@example.com", Language: language},
		EventBeforeUpdate: before,
		EventAfterUpdate:  after,
	}
}

func TestModifiedEventWithoutChange(t *testing.T) {
	location := montreal(t)
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, location)
	before := practice(location)
	after := before
	// Spaces around a text and a pin that barely moves are not changes.
	after.Description = " Ancienne description.\n"
	after.Location.Lat += 0.000001

	if ModifiedEventHasChanges(before, after) {
		t.Error("Expected no change")
	}
	if ModifiedEventNeedsAnswer(before, after) {
		t.Error("Expected no answer to be needed")
	}
	_, changed, err := modifiedEventEmail(modifiedPayload(before, after, "fr"), location, now)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("Expected no email when nothing changed")
	}
}

// Every field a member can see is compared.
func TestModifiedEventHasChanges(t *testing.T) {
	location := montreal(t)
	cases := map[string]func(*model.Event){
		"name":        func(e *model.Event) { e.Name = "Diada" },
		"start":       func(e *model.Event) { e.StartDate += 1800 },
		"end":         func(e *model.Event) { e.EndDate += 1800 },
		"description": func(e *model.Event) { e.Description = "Nouvelle description." },
		"place":       func(e *model.Event) { e.LocationName = "Parc La Fontaine" },
		"map":         func(e *model.Event) { e.Location.Lat += 0.01 },
		"type":        func(e *model.Event) { e.Type = "presentation" },
		"uniform":     func(e *model.Event) { e.UniformRequired = 1 },
		"dues":        func(e *model.Event) { e.DuesCollectionUUID = "d1" },
	}
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, location)
	for name, modify := range cases {
		before := practice(location)
		after := before
		modify(&after)
		if !ModifiedEventHasChanges(before, after) {
			t.Errorf("%s: expected a change", name)
		}
		email, changed, err := modifiedEventEmail(modifiedPayload(before, after, "fr"), location, now)
		if err != nil {
			t.Fatal(err)
		}
		if !changed || strings.TrimSpace(email.MainSections[0].Text) == "" {
			t.Errorf("%s: expected the email to say what changed", name)
		}
		wantAnswer := name == "start" || name == "end" || name == "place"
		if ModifiedEventNeedsAnswer(before, after) != wantAnswer {
			t.Errorf("%s: expected ModifiedEventNeedsAnswer to be %v", name, wantAnswer)
		}
	}
}

func TestModifiedEventSubject(t *testing.T) {
	location := montreal(t)
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, location)
	before := practice(location)

	moved := oneDayLater(before)
	elsewhere := before
	elsewhere.LocationName = "Parc La Fontaine"
	movedElsewhere := moved
	movedElsewhere.LocationName = "Parc La Fontaine"
	described := before
	described.Description = "Nouvelle description."
	described.UniformRequired = 1
	// The date wins over the rest; the new name is the one in the subject.
	movedRenamed := moved
	movedRenamed.Name = "Grand\r\nentraînement"
	movedRenamed.Description = "Nouvelle description."
	halfPast := before
	halfPast.StartDate += 1800

	cases := []struct {
		after    model.Event
		language string
		want     string
	}{
		{moved, "fr", "Changement d'horaire : Entraînement, ven. 20 nov. à 19 h"},
		{elsewhere, "fr", "Changement de lieu : Entraînement, jeu. 19 nov. à 19 h"},
		{movedElsewhere, "fr", "Changement d'horaire et de lieu : Entraînement, ven. 20 nov. à 19 h"},
		{described, "fr", "Évènement mis à jour : Entraînement, jeu. 19 nov. à 19 h"},
		{movedRenamed, "fr", "Changement d'horaire : Grand entraînement, ven. 20 nov. à 19 h"},
		{halfPast, "fr", "Changement d'horaire : Entraînement, jeu. 19 nov. à 19 h 30"},
		{moved, "en", "Schedule change: Entraînement, Fri, Nov 20 at 7 PM"},
		{halfPast, "en", "Schedule change: Entraînement, Thu, Nov 19 at 7:30 PM"},
		{moved, "cat", "Canvi d'horari: Entraînement, dv. 20 nov. a les 19:00"},
	}
	for _, c := range cases {
		email, changed, err := modifiedEventEmail(modifiedPayload(before, c.after, c.language), location, now)
		if err != nil {
			t.Fatal(err)
		}
		if !changed {
			t.Errorf("Expected an email for the subject '%s'", c.want)
			continue
		}
		if email.Header.Title != c.want {
			t.Errorf("Expected the subject '%s'. Got '%s'", c.want, email.Header.Title)
		}
	}
}

// When the date changes, the email says so in one line, shows the event as it
// is now and asks the member to confirm their answer.
func TestModifiedEventDateChange(t *testing.T) {
	location := montreal(t)
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, location)
	before := practice(location)
	payload := modifiedPayload(before, oneDayLater(before), "fr")
	payload.Participation = model.Participation{Answer: common.AnswerYes}
	payload.Token = "token"
	payload.Dependents = []model.Member{{UUID: "m2", FirstName: "Laia"}}

	email, _, err := modifiedEventEmail(payload, location, now)
	if err != nil {
		t.Fatal(err)
	}
	changes := email.MainSections[0].Text
	for _, want := range []string{"Horaire:", "<s style=\"color: #999999;\">jeu. 19 nov., de 19 h à 21 h</s>", "<strong>ven. 20 nov., de 19 h à 21 h</strong>"} {
		if !strings.Contains(changes, want) {
			t.Errorf("Expected the changes to contain '%s'. Got '%s'", want, changes)
		}
	}
	if strings.Count(changes, "<br/>") != 1 {
		t.Errorf("Expected one line of changes. Got '%s'", changes)
	}
	current := email.MainSections[1].Text
	for _, want := range []string{"<strong>Entraînement</strong> (Entraînement)", "ven. 20 nov., de 19 h à 21 h", "Lieu: Casal", "Ancienne description."} {
		if !strings.Contains(current, want) {
			t.Errorf("Expected the event to contain '%s'. Got '%s'", want, current)
		}
	}
	if strings.Contains(current, "jeu. 19 nov.") {
		t.Errorf("Expected the event not to show its old date. Got '%s'", current)
	}
	// The member, the dependent, then the link to the event.
	if len(email.Actions) != 3 {
		t.Fatalf("Expected 3 actions. Got %d", len(email.Actions))
	}
	if !strings.Contains(email.Actions[0].Text, "Je serai présent.") {
		t.Errorf("Expected the answer of the member to be recalled. Got '%s'", email.Actions[0].Text)
	}
	if link := email.Actions[0].Buttons[0].Link; !strings.HasSuffix(link, "/eventShow/e1?a=participate&m=m1&t=token&p=yes") {
		t.Errorf("Unexpected link to answer: '%s'", link)
	}
	if link := email.Actions[1].Buttons[1].Link; !strings.HasSuffix(link, "/eventShow/e1?a=participate&m=m2&t=token&p=no") {
		t.Errorf("Unexpected link to answer for the dependent: '%s'", link)
	}

	// Without a token, there is nothing to answer with.
	payload.Token = ""
	email, _, err = modifiedEventEmail(payload, location, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(email.Actions) != 1 {
		t.Errorf("Expected only the link to the event without a token. Got %d actions", len(email.Actions))
	}
}

// The old description is never quoted: the email only says it changed, and
// shows the new one.
func TestModifiedEventDescriptionChange(t *testing.T) {
	location := montreal(t)
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, location)
	before := practice(location)
	after := before
	after.Description = "Première ligne.\nApporte <ta> gourde & ta faixa."
	payload := modifiedPayload(before, after, "fr")
	payload.Token = "token"

	email, _, err := modifiedEventEmail(payload, location, now)
	if err != nil {
		t.Fatal(err)
	}
	changes := email.MainSections[0].Text
	if !strings.Contains(changes, "La description a été mise à jour.") {
		t.Errorf("Expected the changes to mention the description. Got '%s'", changes)
	}
	current := email.MainSections[1].Text
	if !strings.Contains(current, "Première ligne.<br/>Apporte &lt;ta&gt; gourde &amp; ta faixa.") {
		t.Errorf("Expected the new description, escaped, with its line break. Got '%s'", current)
	}
	if strings.Contains(changes+current, "Ancienne description.") {
		t.Error("Expected the old description to be left out")
	}
	// Nobody has to answer again for a new description.
	if len(email.Actions) != 1 {
		t.Errorf("Expected only the link to the event. Got %d actions", len(email.Actions))
	}

	// The whole email can be built, and the event is not escaped twice.
	body, err := email.buildEmail()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "Apporte &lt;ta&gt; gourde &amp; ta faixa.") {
		t.Error("Expected the description in the email")
	}
	if !strings.Contains(body, "Subject: Évènement mis à jour : Entraînement, jeu. 19 nov. à 19 h\r\n") {
		t.Error("Expected the subject in the headers of the email")
	}
}

// The uniform, the dues, the type and the place are all told.
func TestModifiedEventOtherChanges(t *testing.T) {
	location := montreal(t)
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, location)
	before := practice(location)
	before.LocationName = ""
	after := before
	after.LocationName = "Parc La Fontaine"
	after.Type = "presentation"
	after.UniformRequired = 1
	after.DuesCollectionUUID = "d1"
	payload := modifiedPayload(before, after, "fr")
	payload.DuesAfter = &DuesReminder{Kind: "winter", Amount: 2000}

	email, _, err := modifiedEventEmail(payload, location, now)
	if err != nil {
		t.Fatal(err)
	}
	changes := email.MainSections[0].Text
	for _, want := range []string{
		"<strong>Lieu:</strong> <s style=\"color: #999999;\">—</s> &#x2192; <strong>Parc La Fontaine</strong>",
		"<s style=\"color: #999999;\">Entraînement</s> &#x2192; <strong>Présentation</strong>",
		"L&#39;uniforme est maintenant de mise.",
		"Une cotisation sera recueillie lors de cet évènement : Cotisation d&#39;hiver (20 $).",
	} {
		if !strings.Contains(changes, want) {
			t.Errorf("Expected the changes to contain '%s'. Got '%s'", want, changes)
		}
	}
	current := email.MainSections[1].Text
	for _, want := range []string{"Lieu: Parc La Fontaine", "Uniforme de mise", "Cotisation recueillie: Cotisation d&#39;hiver (20 $)"} {
		if !strings.Contains(current, want) {
			t.Errorf("Expected the event to contain '%s'. Got '%s'", want, current)
		}
	}

	// The dues are removed, the uniform too.
	payload = modifiedPayload(after, before, "fr")
	payload.DuesBefore = &DuesReminder{Kind: "winter", Amount: 2000}
	email, _, err = modifiedEventEmail(payload, location, now)
	if err != nil {
		t.Fatal(err)
	}
	changes = email.MainSections[0].Text
	for _, want := range []string{"L&#39;uniforme n&#39;est plus requis.", "Il n&#39;y a plus de cueillette de cotisation lors de cet évènement."} {
		if !strings.Contains(changes, want) {
			t.Errorf("Expected the changes to contain '%s'. Got '%s'", want, changes)
		}
	}
}

func TestFormatWhen(t *testing.T) {
	location := montreal(t)
	now := time.Date(2026, time.November, 17, 12, 0, 0, 0, location)
	// A party that ends after midnight, next year.
	party := model.Event{
		StartDate: uint(time.Date(2027, time.January, 9, 20, 30, 0, 0, location).Unix()),
		EndDate:   uint(time.Date(2027, time.January, 10, 1, 0, 0, 0, location).Unix()),
	}
	cases := map[string]string{
		"fr":  "sam. 9 janv. 2027 à 20 h 30 – dim. 10 janv. 2027 à 1 h",
		"en":  "Sat, Jan 9, 2027 at 8:30 PM – Sun, Jan 10, 2027 at 1 AM",
		"cat": "ds. 9 gen. 2027 a les 20:30 – dg. 10 gen. 2027 a les 01:00",
		// An unknown language is written in French.
		"": "sam. 9 janv. 2027 à 20 h 30 – dim. 10 janv. 2027 à 1 h",
	}
	for language, want := range cases {
		if got := formatWhen(party, location, now, language); got != want {
			t.Errorf("%s: expected '%s'. Got '%s'", language, want, got)
		}
	}
	if got, want := formatWhen(practice(location), location, now, "en"), "Thu, Nov 19, from 7 PM to 9 PM"; got != want {
		t.Errorf("Expected '%s'. Got '%s'", want, got)
	}
}

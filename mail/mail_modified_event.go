package mail

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

// EmailModifiedPayload is what the email about a modified event needs. The
// notification only stores the two versions of the event; the rest is filled
// for each member when the email is sent.
type EmailModifiedPayload struct {
	Member            model.Member `json:"member"`
	EventBeforeUpdate model.Event  `json:"eventBeforeUpdate"`
	EventAfterUpdate  model.Event  `json:"eventAfterUpdate"`
	// DuesBefore and DuesAfter describe the dues collected at the event, before
	// and after the update. Nil when there are none, or when they are unknown.
	DuesBefore *DuesReminder `json:"duesBefore,omitempty"`
	DuesAfter  *DuesReminder `json:"duesAfter,omitempty"`
	// Participation, Token and Dependents are only needed when the member is
	// asked to confirm their answer (see ModifiedEventNeedsAnswer). Without a
	// token, the email has no answer buttons.
	Participation model.Participation `json:"participation"`
	Token         string              `json:"token,omitempty"`
	Dependents    []model.Member      `json:"dependents,omitempty"`
}

// change is one line of the "what changed" box: either a short value that
// moved from Before to After, or a Sentence that says it all.
type change struct {
	Label    string
	Before   string
	After    string
	Sentence string
}

// eventChanges is everything that differs between two versions of an event.
type eventChanges struct {
	// When and Where are the changes a member must not miss: they decide the
	// subject of the email and whether the member is asked to confirm.
	When  bool
	Where bool
	Lines []change
}

// A map pin can move by this many degrees (about a metre) without counting as
// a change: coordinates do not always survive a round trip to the database.
const sameLocationTolerance = 0.00001

func sameText(a, b string) bool {
	return strings.TrimSpace(a) == strings.TrimSpace(b)
}

func whenChanged(before, after model.Event) bool {
	return before.StartDate != after.StartDate || before.EndDate != after.EndDate
}

func whereChanged(before, after model.Event) bool {
	return !sameText(before.LocationName, after.LocationName)
}

func mapPinMoved(before, after model.Event) bool {
	return math.Abs(before.Location.Lat-after.Location.Lat) > sameLocationTolerance ||
		math.Abs(before.Location.Lng-after.Location.Lng) > sameLocationTolerance
}

// ModifiedEventHasChanges tells if an update changed anything a member can
// see. When it did not, there is nothing to tell the members.
func ModifiedEventHasChanges(before, after model.Event) bool {
	return whenChanged(before, after) ||
		whereChanged(before, after) ||
		mapPinMoved(before, after) ||
		!sameText(before.Name, after.Name) ||
		!sameText(before.Description, after.Description) ||
		before.Type != after.Type ||
		before.UniformRequired != after.UniformRequired ||
		before.DuesCollectionUUID != after.DuesCollectionUUID
}

// ModifiedEventNeedsAnswer tells if the update can change who is able to come:
// the date, the time or the place moved. The members are then asked to confirm
// their answer.
func ModifiedEventNeedsAnswer(before, after model.Event) bool {
	return whenChanged(before, after) || whereChanged(before, after)
}

// dateWords holds what is needed to write a date in a language.
type dateWords struct {
	weekdays [7]string // Sunday first
	months   [12]string
	at       string
	from     string
	to       string
}

var dateWordsByLanguage = map[string]dateWords{
	"fr": {
		weekdays: [7]string{"dim.", "lun.", "mar.", "mer.", "jeu.", "ven.", "sam."},
		months:   [12]string{"janv.", "févr.", "mars", "avr.", "mai", "juin", "juill.", "août", "sept.", "oct.", "nov.", "déc."},
		at:       "à", from: "de", to: "à",
	},
	"en": {
		weekdays: [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
		months:   [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
		at:       "at", from: "from", to: "to",
	},
	"ca": {
		weekdays: [7]string{"dg.", "dl.", "dt.", "dc.", "dj.", "dv.", "ds."},
		months:   [12]string{"gen.", "febr.", "març", "abr.", "maig", "juny", "jul.", "ag.", "set.", "oct.", "nov.", "des."},
		at:       "a les", from: "de", to: "a",
	},
}

// dateLanguage gives the language dates are written in for a member: French
// unless the member chose English or Catalan (stored as "cat").
func dateLanguage(language string) string {
	switch language {
	case "en":
		return "en"
	case "cat", "ca":
		return "ca"
	}
	return "fr"
}

// formatDay writes "ven. 20 nov." (fr), "Fri, Nov 20" (en) or "dv. 20 nov."
// (ca). The year is only written when it is not the current one.
func formatDay(t, now time.Time, language string) string {
	language = dateLanguage(language)
	words := dateWordsByLanguage[language]
	var day string
	if language == "en" {
		day = fmt.Sprintf("%s, %s %d", words.weekdays[t.Weekday()], words.months[t.Month()-1], t.Day())
		if t.Year() != now.Year() {
			day += fmt.Sprintf(", %d", t.Year())
		}
		return day
	}
	day = fmt.Sprintf("%s %d %s", words.weekdays[t.Weekday()], t.Day(), words.months[t.Month()-1])
	if t.Year() != now.Year() {
		day += fmt.Sprintf(" %d", t.Year())
	}
	return day
}

// formatTime writes "19 h" or "19 h 30" (fr), "7 PM" or "7:30 PM" (en),
// "19:00" (ca).
func formatTime(t time.Time, language string) string {
	switch dateLanguage(language) {
	case "en":
		if t.Minute() == 0 {
			return t.Format("3 PM")
		}
		return t.Format("3:04 PM")
	case "ca":
		return t.Format("15:04")
	}
	if t.Minute() == 0 {
		return fmt.Sprintf("%d h", t.Hour())
	}
	return fmt.Sprintf("%d h %02d", t.Hour(), t.Minute())
}

// formatStart writes when an event starts: "ven. 20 nov. à 19 h".
func formatStart(event model.Event, location *time.Location, now time.Time, language string) string {
	start := time.Unix(int64(event.StartDate), 0).In(location)
	words := dateWordsByLanguage[dateLanguage(language)]
	return formatDay(start, now, language) + " " + words.at + " " + formatTime(start, language)
}

// formatWhen writes when an event takes place: "ven. 20 nov., de 19 h à 21 h",
// or both ends in full when it does not end the day it starts.
func formatWhen(event model.Event, location *time.Location, now time.Time, language string) string {
	start := time.Unix(int64(event.StartDate), 0).In(location)
	end := time.Unix(int64(event.EndDate), 0).In(location)
	words := dateWordsByLanguage[dateLanguage(language)]
	if start.Year() == end.Year() && start.YearDay() == end.YearDay() {
		return formatDay(start, now, language) + ", " +
			words.from + " " + formatTime(start, language) + " " +
			words.to + " " + formatTime(end, language)
	}
	return formatDay(start, now, language) + " " + words.at + " " + formatTime(start, language) +
		" – " +
		formatDay(end, now, language) + " " + words.at + " " + formatTime(end, language)
}

// duesLabel writes the dues collected at an event: "Cotisation d'hiver (20 $)".
func duesLabel(dues DuesReminder, language string) string {
	return fmt.Sprintf("%s (%s)", common.Translate("dues_kind_"+dues.Kind, language), formatDuesAmount(dues.Amount, language))
}

// eventTypeLabel translates the type of an event; an event without a type has
// no label.
func eventTypeLabel(eventType, language string) string {
	if eventType == "" {
		return ""
	}
	return common.Translate(eventType, language)
}

// orDash stands for a value that was, or became, empty.
func orDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return strings.TrimSpace(value)
}

// modifiedEventChanges lists what changed, the most important first. Long
// texts are never quoted: the email shows the event as it is now.
func modifiedEventChanges(payload EmailModifiedPayload, location *time.Location, now time.Time) eventChanges {
	before, after := payload.EventBeforeUpdate, payload.EventAfterUpdate
	language := payload.Member.Language
	changes := eventChanges{When: whenChanged(before, after), Where: whereChanged(before, after)}

	if changes.When {
		changes.Lines = append(changes.Lines, change{
			Label:  common.Translate("modified_event_when", language),
			Before: formatWhen(before, location, now, language),
			After:  formatWhen(after, location, now, language),
		})
	}
	if changes.Where {
		changes.Lines = append(changes.Lines, change{
			Label:  common.Translate("modified_event_location", language),
			Before: orDash(before.LocationName),
			After:  orDash(after.LocationName),
		})
	} else if mapPinMoved(before, after) {
		changes.Lines = append(changes.Lines, change{Sentence: common.Translate("modified_event_map_moved", language)})
	}
	if !sameText(before.Name, after.Name) {
		changes.Lines = append(changes.Lines, change{
			Label:  common.Translate("modified_event_name", language),
			Before: orDash(before.Name),
			After:  orDash(after.Name),
		})
	}
	if before.Type != after.Type {
		changes.Lines = append(changes.Lines, change{
			Label:  common.Translate("modified_event_type", language),
			Before: orDash(eventTypeLabel(before.Type, language)),
			After:  orDash(eventTypeLabel(after.Type, language)),
		})
	}
	if before.UniformRequired != after.UniformRequired {
		key := "modified_event_uniform_removed"
		if after.UniformRequired == 1 {
			key = "modified_event_uniform_added"
		}
		changes.Lines = append(changes.Lines, change{Sentence: common.Translate(key, language)})
	}
	if before.DuesCollectionUUID != after.DuesCollectionUUID {
		switch {
		case after.DuesCollectionUUID == "":
			changes.Lines = append(changes.Lines, change{Sentence: common.Translate("modified_event_dues_removed", language)})
		case before.DuesCollectionUUID == "" && payload.DuesAfter != nil:
			changes.Lines = append(changes.Lines, change{
				Sentence: fmt.Sprintf(common.Translate("modified_event_dues_added", language), duesLabel(*payload.DuesAfter, language)),
			})
		case payload.DuesBefore != nil && payload.DuesAfter != nil:
			changes.Lines = append(changes.Lines, change{
				Label:  common.Translate("modified_event_dues", language),
				Before: duesLabel(*payload.DuesBefore, language),
				After:  duesLabel(*payload.DuesAfter, language),
			})
		default:
			// The dues changed but we could not find out what they are.
			changes.Lines = append(changes.Lines, change{Sentence: common.Translate("modified_event_dues_changed", language)})
		}
	}
	if !sameText(before.Description, after.Description) {
		changes.Lines = append(changes.Lines, change{Sentence: common.Translate("modified_event_description_updated", language)})
	}
	return changes
}

// singleLine keeps a value on one line, so that it can go in a mail header.
func singleLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

// modifiedEventSubject writes the subject of the email. It names the change
// that matters most (the date or time, then the place), and always ends with
// the name and the start of the event as they are now.
func modifiedEventSubject(changes eventChanges, after model.Event, location *time.Location, now time.Time, language string) string {
	key := "modified_event_subject_other"
	switch {
	case changes.When && changes.Where:
		key = "modified_event_subject_when_where"
	case changes.When:
		key = "modified_event_subject_when"
	case changes.Where:
		key = "modified_event_subject_where"
	}
	return singleLine(common.Translate(key, language) + " " + after.Name + ", " + formatStart(after, location, now, language))
}

func changesString(changes []change) (string, error) {
	const templateChanges = `{{ range . }}{{ if .Sentence }}{{ .Sentence }}{{ else }}<strong>{{ .Label }}:</strong> <s style="color: #999999;">{{ .Before }}</s> &#x2192; <strong>{{ .After }}</strong>{{ end }}<br/>
{{ end }}`
	t, err := template.New("changes").Parse(templateChanges)
	if err != nil {
		common.Error("Error parsing template: " + err.Error())
		return "", err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, changes); err != nil {
		common.Error("Error generating template: " + err.Error())
		return "", err
	}
	return buffer.String(), nil
}

// currentEvent is the event as it is after the update, ready to be shown.
type currentEvent struct {
	Name          string
	Type          string
	When          string
	LocationLabel string
	Location      string
	Uniform       string
	Dues          string
	Description   template.HTML
}

// currentEventString writes the event as it is now: this is what the member
// should rely on, whatever changed.
func currentEventString(payload EmailModifiedPayload, location *time.Location, now time.Time) (string, error) {
	event := payload.EventAfterUpdate
	language := payload.Member.Language
	current := currentEvent{
		Name:          event.Name,
		Type:          eventTypeLabel(event.Type, language),
		When:          formatWhen(event, location, now, language),
		LocationLabel: common.Translate("create_event_main_location", language),
		Location:      strings.TrimSpace(event.LocationName),
	}
	if event.UniformRequired == 1 {
		current.Uniform = common.Translate("reminder_uniform_title", language)
	}
	if event.DuesCollectionUUID != "" && payload.DuesAfter != nil {
		current.Dues = common.Translate("modified_event_dues", language) + ": " + duesLabel(*payload.DuesAfter, language)
	}
	// The description is plain text: escape it, then keep its line breaks.
	description := template.HTMLEscapeString(strings.TrimSpace(event.Description))
	description = strings.ReplaceAll(description, "\r\n", "\n")
	current.Description = template.HTML(strings.ReplaceAll(description, "\n", "<br/>"))

	const templateEvent = `<strong>{{ .Name }}</strong>{{ if .Type }} ({{ .Type }}){{ end }}<br/>
{{ .When }}<br/>
{{ if .Location }}{{ .LocationLabel }} {{ .Location }}<br/>
{{ end }}{{ if .Uniform }}{{ .Uniform }}<br/>
{{ end }}{{ if .Dues }}{{ .Dues }}<br/>
{{ end }}{{ if .Description }}<br/>
{{ .Description }}<br/>
{{ end }}`
	t, err := template.New("event").Parse(templateEvent)
	if err != nil {
		common.Error("Error parsing template: " + err.Error())
		return "", err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, current); err != nil {
		common.Error("Error generating template: " + err.Error())
		return "", err
	}
	return buffer.String(), nil
}

// answerButtons are the yes and no buttons a member answers with, for
// themselves or for a dependent.
func answerButtons(eventUUID, memberUUID, token, yesKey, noKey, language string) []Button {
	participationLink := common.GetConfigString("domain") + "/eventShow/" + eventUUID + "?" +
		"a=participate" +
		"&m=" + memberUUID +
		"&t=" + token +
		"&p="
	return []Button{
		{Text: common.Translate(yesKey, language), Link: participationLink + "yes", Color: "#20470b"},
		{Text: common.Translate(noKey, language), Link: participationLink + "no", Color: "#aa0000"},
	}
}

// confirmActions asks the member, and each of their dependents, to confirm
// their answer now that the date, the time or the place changed.
func confirmActions(payload EmailModifiedPayload) []emailAction {
	language := payload.Member.Language
	event := payload.EventAfterUpdate
	text := common.Translate("modified_event_confirm_unanswered", language)
	if payload.Participation.Answer == common.AnswerYes || payload.Participation.Answer == common.AnswerNo {
		text = fmt.Sprintf(common.Translate("modified_event_confirm_answered", language),
			common.Translate("reminder_answer_"+payload.Participation.Answer, language))
	}
	actions := []emailAction{{
		Title:   common.Translate("reminder_availability", language),
		Text:    text,
		Buttons: answerButtons(event.UUID, payload.Member.UUID, payload.Token, "reminder_answer_yes", "reminder_answer_no", language),
	}}
	for _, dependent := range payload.Dependents {
		actions = append(actions, emailAction{
			Title:   common.Translate("reminder_availability_dependent", language) + dependent.FirstName,
			Text:    fmt.Sprintf(common.Translate("reminder_confirm_dependent", language), dependent.FirstName),
			Buttons: answerButtons(event.UUID, dependent.UUID, payload.Token, "reminder_answer_yes_dependent", "reminder_answer_no_dependent", language),
		})
	}
	return actions
}

// modifiedEventEmail builds the email. The second value is false when the two
// versions of the event show no difference: there is then nothing to send.
func modifiedEventEmail(payload EmailModifiedPayload, location *time.Location, now time.Time) (emailInfo, bool, error) {
	language := payload.Member.Language
	changes := modifiedEventChanges(payload, location, now)
	if len(changes.Lines) == 0 {
		return emailInfo{}, false, nil
	}
	changesText, err := changesString(changes.Lines)
	if err != nil {
		return emailInfo{}, false, err
	}
	eventText, err := currentEventString(payload, location, now)
	if err != nil {
		return emailInfo{}, false, err
	}

	email := emailInfo{}
	email.Header = emailHeader{Title: modifiedEventSubject(changes, payload.EventAfterUpdate, location, now, language)}
	email.Top = emailTop{
		Title:    common.Translate("greetings", language) + " " + payload.Member.FirstName,
		Subtitle: common.Translate("modified_event_intro", language),
		To:       payload.Member.Email,
	}
	email.MainSections = []emailMain{
		{Title: common.Translate("modified_event_changes_title", language), Text: changesText},
		{Title: common.Translate("modified_event_current_title", language), Text: eventText},
	}
	if (changes.When || changes.Where) && payload.Token != "" {
		email.Actions = confirmActions(payload)
	}
	email.Actions = append(email.Actions, emailAction{
		Title: common.Translate("modified_event_action_title", language),
		Text:  common.Translate("modified_event_action_text", language),
		Buttons: []Button{{
			Text: common.Translate("modified_event_action_button", language),
			Link: common.GetConfigString("domain") + "/eventShow/" + payload.EventAfterUpdate.UUID,
		}},
	})
	email.Bottom = emailBottom{
		ProfileLink: common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID,
		MyProfile:   common.Translate("email_my_profile", language),
		Suggestions: common.Translate("email_suggestions", language),
	}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"
	return email, true, nil
}

// SendModifiedEventEmail tells a member that an event changed: what changed,
// then the event as it is now. Nothing is sent when nothing visible changed.
func SendModifiedEventEmail(ctx context.Context, payload EmailModifiedPayload) error {
	ctx, span := tracer.Start(ctx, "mail.SendModifiedEventEmail")
	defer span.End()

	location, err := time.LoadLocation("America/Montreal")
	if err != nil {
		common.Error("%v\n", err)
		return err
	}
	email, changed, err := modifiedEventEmail(payload, location, time.Now().In(location))
	if err != nil {
		return err
	}
	if !changed {
		common.Debug("Event %s was updated without visible change, not sending the email", payload.EventAfterUpdate.UUID)
		return nil
	}
	if err = sendMail(ctx, email); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

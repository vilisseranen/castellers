package mail

import (
	"bytes"
	"context"
	"html/template"
	"strings"
	"time"

	"github.com/tommysolsen/capitalise"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

type EmailCreateEventPayload struct {
	Member model.Member `json:"member"`
	// EventBeforeUpdate model.Event  `json:"eventBeforeUpdate"`
	Event model.Event `json:"event"`
}

type eventDetails struct {
	Name        string
	Type        string
	Date        string
	Location    string
	Description string
	Recurring   string
}

func SendCreateEventEmail(ctx context.Context, payload EmailCreateEventPayload) error {
	ctx, span := tracer.Start(ctx, "mail.SendCreateEventEmail")
	defer span.End()

	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID
	location, err := time.LoadLocation("America/Montreal")
	if err != nil {
		common.Error("%v\n", err)
		return err
	}
	eventDate := time.Unix(int64(payload.Event.StartDate), 0).In(location).Format("02-01-2006")

	email := emailInfo{}
	email.Header = emailHeader{common.Translate("create_event_subject", payload.Member.Language)}
	email.Top = emailTop{
		Title:    common.Translate("greetings", payload.Member.Language) + " " + payload.Member.FirstName,
		Subtitle: common.Translate("create_event_intro", payload.Member.Language),
		To:       payload.Member.Email,
	}
	eventDetails, err := eventDetailsString(payload.Event, payload.Member, location)
	if err != nil {
		return err
	}
	email.MainSections = []emailMain{{
		Title:    payload.Event.Name + " " + common.Translate("on_the", payload.Member.Language) + " " + eventDate + ".",
		Subtitle: common.Translate("create_event_text", payload.Member.Language),
		Text:     eventDetails,
	}}
	email.Action = emailAction{
		Title: common.Translate("create_event_action_title", payload.Member.Language),
		Text:  common.Translate("create_event_action_text", payload.Member.Language),
		Buttons: []Button{{
			Text: common.Translate("create_event_action_button", payload.Member.Language),
			Link: common.GetConfigString("domain") + "/eventShow/" + payload.Event.UUID,
		}},
	}
	email.Bottom = emailBottom{ProfileLink: profileLink, MyProfile: common.Translate("email_my_profile", payload.Member.Language), Suggestions: common.Translate("email_suggestions", payload.Member.Language)}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"

	if err = sendMail(ctx, email); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

func eventDetailsString(event model.Event, member model.Member, location *time.Location) (string, error) {
	eventDetails := eventDetails{Name: event.Name, Type: common.Translate(event.Type, member.Language), Description: event.Description}
	eventDate := time.Unix(int64(event.StartDate), 0).In(location).Format("02-01-2006")
	eventStartTime := time.Unix(int64(event.StartDate), 0).In(location).Format("15:04")
	eventEndTime := time.Unix(int64(event.EndDate), 0).In(location).Format("15:04")
	eventDetails.Date = strings.Join(
		[]string{
			capitalise.First(common.Translate("on_the", member.Language)),
			eventDate,
			common.Translate("from_time", member.Language),
			eventStartTime,
			common.Translate("to_time", member.Language),
			eventEndTime,
		}, " ")
	eventDetails.Date += "."
	eventDetails.Location = common.Translate("create_event_main_location", member.Language) + " " + event.LocationName + "."
	if event.Recurring.Until != 0 {
		eventDetails.Recurring = common.Translate("create_event_main_recurring", member.Language)
	}
	const templateChanges = `
	<strong>{{ .Name }}</strong> ({{ .Type }})<br/><br/>
	{{ .Date }}<br/>
	{{ .Description}}<br/><br/>
	{{ if .Recurring }} {{ .Recurring }} <br/><br/>{{ end }}
	`
	t, err := template.New("event").Parse(templateChanges)
	if err != nil {
		common.Error("Error parsing template: " + err.Error())
		return "", err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, eventDetails); err != nil {
		common.Error("Error generating template: " + err.Error())
		return "", err
	}
	return buffer.String(), nil
}

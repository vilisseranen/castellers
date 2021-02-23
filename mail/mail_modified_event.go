package mail

import (
	"bytes"
	"text/template"
	"time"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

type EmailModifiedPayload struct {
	Member            model.Member `json:"member"`
	EventBeforeUpdate model.Event  `json:"eventBeforeUpdate"`
	EventAfterUpdate  model.Event  `json:"eventAfterUpdate"`
}

type change struct {
	Type   string
	Before string
	After  string
}

func SendModifiedEventEmail(payload EmailModifiedPayload) error {

	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID
	location, err := time.LoadLocation("America/Montreal")
	if err != nil {
		common.Error("%v\n", err)
		return err
	}
	eventDate := time.Unix(int64(payload.EventAfterUpdate.StartDate), 0).In(location).Format("02-01-2006")

	// Calculate changes
	var changes []change
	if payload.EventBeforeUpdate.Name != payload.EventAfterUpdate.Name {
		change := change{Type: common.Translate("modified_event_name", payload.Member.Language),
			Before: payload.EventBeforeUpdate.Name, After: payload.EventAfterUpdate.Name}
		changes = append(changes, change)
	}
	if payload.EventBeforeUpdate.StartDate != payload.EventAfterUpdate.StartDate {
		change := change{
			Type:   common.Translate("modified_event_start_date", payload.Member.Language),
			Before: time.Unix(int64(payload.EventBeforeUpdate.StartDate), 0).In(location).Format("02-01-2006 15:04"),
			After:  time.Unix(int64(payload.EventAfterUpdate.StartDate), 0).In(location).Format("02-01-2006 15:04"),
		}
		changes = append(changes, change)
	}
	if payload.EventBeforeUpdate.EndDate != payload.EventAfterUpdate.EndDate {
		change := change{
			Type:   common.Translate("modified_event_end_date", payload.Member.Language),
			Before: time.Unix(int64(payload.EventBeforeUpdate.EndDate), 0).In(location).Format("02-01-2006 15:04"),
			After:  time.Unix(int64(payload.EventAfterUpdate.EndDate), 0).In(location).Format("02-01-2006 15:04"),
		}
		changes = append(changes, change)
	}
	if payload.EventBeforeUpdate.Description != payload.EventAfterUpdate.Description {
		change := change{Type: common.Translate("modified_event_description", payload.Member.Language),
			Before: payload.EventBeforeUpdate.Description, After: payload.EventAfterUpdate.Description}
		changes = append(changes, change)
	}
	if payload.EventBeforeUpdate.LocationName != payload.EventAfterUpdate.LocationName {
		change := change{Type: common.Translate("modified_event_location", payload.Member.Language),
			Before: payload.EventBeforeUpdate.LocationName, After: payload.EventAfterUpdate.LocationName}
		changes = append(changes, change)
	}

	email := emailInfo{}
	email.Header = emailHeader{common.Translate("modified_event_subject", payload.Member.Language)}
	email.Top = emailTop{
		Title:    common.Translate("greetings", payload.Member.Language) + " " + payload.Member.FirstName,
		Subtitle: common.Translate("modified_event_intro", payload.Member.Language),
		To:       payload.Member.Email,
	}
	changesString, err := changesString(changes)
	if err != nil {
		return err
	}
	email.MainSections = []emailMain{{
		Title:    payload.EventAfterUpdate.Name + " " + common.Translate("on_the", payload.Member.Language) + " " + eventDate + ".",
		Subtitle: common.Translate("modified_event_text", payload.Member.Language),
		Text:     changesString,
	}}
	email.Action = emailAction{
		Title: common.Translate("modified_event_action_title", payload.Member.Language),
		Text:  common.Translate("modified_event_action_text", payload.Member.Language),
		Buttons: []Button{{
			Text: common.Translate("modified_event_action_button", payload.Member.Language),
			Link: common.GetConfigString("domain") + "/eventShow/" + payload.EventAfterUpdate.UUID,
		}},
	}
	email.Bottom = emailBottom{ProfileLink: profileLink, MyProfile: common.Translate("email_my_profile", payload.Member.Language), Suggestions: common.Translate("email_suggestions", payload.Member.Language)}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"

	if err = sendMail(email); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

func changesString(changes []change) (string, error) {
	const templateChanges = `
	{{ range . }}
	  <strong>{{ .Type }}:</strong> {{ .Before }} &#x2192; {{ .After}}
	  <br/>
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

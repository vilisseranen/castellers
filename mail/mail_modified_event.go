package mail

import (
	"bytes"
	"html/template"
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

type emailModifiedEventInfo struct {
	Subject            string
	Greetings          string
	MemberName         string
	ParticipationLink  string
	ImageSource        string
	EventFormatted     string
	ModifiedEventIntro string
	ModifiedEventText  string
	Changes            []change
}

func (e emailModifiedEventInfo) GetBody() (string, error) {
	t, err := template.ParseFiles("mail/templates/email_modified_event_body.html")
	if err != nil {
		common.Error("Error parsing template: " + err.Error())
		return "", err
	}
	body := new(bytes.Buffer)
	if err = t.Execute(body, e); err != nil {
		common.Error("Error generating template: " + err.Error())
		return "", err
	}
	return body.String(), nil
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
	email.Top = emailTop{Title: common.Translate("modified_event_subject", payload.Member.Language), To: payload.Member.Email}
	email.Body = emailModifiedEventInfo{
		Subject:            common.Translate("modified_event_subject", payload.Member.Language),
		Greetings:          common.Translate("modified_event_greetings", payload.Member.Language),
		MemberName:         payload.Member.FirstName,
		ImageSource:        common.GetConfigString("cdn") + "/static/img/",
		EventFormatted:     payload.EventAfterUpdate.Name + " " + common.Translate("reminder_on_the", payload.Member.Language) + " " + eventDate + ".",
		ModifiedEventIntro: common.Translate("modified_event_intro", payload.Member.Language),
		ModifiedEventText:  common.Translate("modified_event_text", payload.Member.Language),
		Changes:            changes,
	}
	email.Bottom = emailBottom{ProfileLink: profileLink, MyProfile: common.Translate("email_my_profile", payload.Member.Language), Suggestions: common.Translate("email_suggestions", payload.Member.Language)}

	emailBodyString, err := email.buildEmail()
	if err != nil {
		return err
	}
	emailString := emailBodyString
	// Send mail
	if err = sendMail([]string{payload.Member.Email}, emailString); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

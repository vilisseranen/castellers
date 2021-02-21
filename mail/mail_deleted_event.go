package mail

import (
	"bytes"
	"html/template"
	"time"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

type EmailDeletedEventPayload struct {
	Member       model.Member `json:"member"`
	EventDeleted model.Event  `json:"eventDeleted"`
}
type emailDeletedEventInfo struct {
	Subject           string
	Greetings         string
	MemberName        string
	ParticipationLink string
	ImageSource       string
	EventFormatted    string
	DeletedEventIntro string
	DeletedEventText  string
}

func (e emailDeletedEventInfo) GetBody() (string, error) {
	t, err := template.ParseFiles("mail/templates/email_deleted_event_body.html")
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

func SendDeletedEventEmail(payload EmailDeletedEventPayload) error {

	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID
	location, err := time.LoadLocation("America/Montreal")
	if err != nil {
		common.Error("%v\n", err)
		return err
	}
	eventDate := time.Unix(int64(payload.EventDeleted.StartDate), 0).In(location).Format("02-01-2006")

	email := emailInfo{}
	email.Top = emailTop{Title: common.Translate("deleted_event_subject", payload.Member.Language), To: payload.Member.Email}
	email.Body = emailDeletedEventInfo{
		Subject:           common.Translate("deleted_event_subject", payload.Member.Language),
		Greetings:         common.Translate("deleted_event_greetings", payload.Member.Language),
		MemberName:        payload.Member.FirstName,
		ImageSource:       common.GetConfigString("cdn") + "/static/img/",
		EventFormatted:    payload.EventDeleted.Name + " " + common.Translate("reminder_on_the", payload.Member.Language) + " " + eventDate + ".",
		DeletedEventIntro: common.Translate("deleted_event_intro", payload.Member.Language),
		DeletedEventText:  common.Translate("deleted_event_text", payload.Member.Language),
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

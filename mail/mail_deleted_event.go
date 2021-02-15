package mail

import (
	"bytes"
	"html/template"

	"github.com/vilisseranen/castellers/common"
)

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

func SendDeletedEventEmail(to, memberName, languageUser, profileLink, eventName, eventDate string) error {
	email := emailInfo{}
	email.Top = emailTop{Title: common.Translate("deleted_event_subject", languageUser), To: to}
	email.Body = emailDeletedEventInfo{
		Subject:           common.Translate("deleted_event_subject", languageUser),
		Greetings:         common.Translate("deleted_event_greetings", languageUser),
		MemberName:        memberName,
		ImageSource:       common.GetConfigString("cdn") + "/static/img/",
		EventFormatted:    eventName + " " + common.Translate("reminder_on_the", languageUser) + " " + eventDate + ".",
		DeletedEventIntro: common.Translate("deleted_event_intro", languageUser),
		DeletedEventText:  common.Translate("deleted_event_text", languageUser),
	}
	email.Bottom = emailBottom{ProfileLink: profileLink, MyProfile: common.Translate("email_my_profile", languageUser), Suggestions: common.Translate("email_suggestions", languageUser)}

	emailBodyString, err := email.buildEmail()
	if err != nil {
		return err
	}
	emailString := emailBodyString
	// Send mail
	if err = sendMail([]string{to}, emailString); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

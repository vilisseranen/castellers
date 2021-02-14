package mail

import (
	"bytes"
	"html/template"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

type emailSummaryInfo struct {
	Subject        string
	Greetings      string
	Inscriptions   string
	EventFormatted string
	FirstName      string
	Name           string
	Roles          string
	Answer         string
	ParticipateYes string
	ParticipateNo  string
	NoAnswer       string
	MemberName     string
	ImageSource    string
	Members        []model.Member
}

func (e emailSummaryInfo) GetBody() (string, error) {
	t, err := template.ParseFiles("mail/templates/email_summary_body.html")
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

func SendSummaryEmail(to, memberName, languageUser, profileLink, eventName, eventDate string, members []model.Member) error {
	email := emailInfo{}
	email.Top = emailTop{Title: common.Translate("summary_subject", languageUser), To: to}
	email.Body = emailSummaryInfo{
		Subject:        common.Translate("summary_subject", languageUser),
		Greetings:      common.Translate("summary_greetings", languageUser),
		Inscriptions:   common.Translate("summary_inscriptions", languageUser),
		EventFormatted: eventName + " " + common.Translate("reminder_on_the", languageUser) + " " + eventDate + ".",
		FirstName:      common.Translate("summary_first_name", languageUser),
		Name:           common.Translate("summary_name", languageUser),
		Roles:          common.Translate("summary_roles", languageUser),
		Answer:         common.Translate("summary_answer", languageUser),
		ParticipateYes: common.Translate("summary_participate_yes", languageUser),
		ParticipateNo:  common.Translate("summary_participate_no", languageUser),
		NoAnswer:       common.Translate("summary_no_answer", languageUser),
		MemberName:     memberName,
		ImageSource:    common.GetConfigString("cdn") + "/static/img/",
		Members:        members,
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

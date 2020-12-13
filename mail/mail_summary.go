package mail

import (
	"bytes"
	"html/template"
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
	// Members              []model.Member // TO FIX
	Members []string
}

func (e emailSummaryInfo) GetBody() (string, error) {
	t, err := template.ParseFiles("templates/email_summary_body.html")
	if err != nil {
		Error("Error parsing template: " + err.Error())
		return "", err
	}
	body := new(bytes.Buffer)
	if err = t.Execute(body, e); err != nil {
		Error("Error generating template: " + err.Error())
		return "", err
	}
	return body.String(), nil
}

func SendSummaryEmail(to, memberName, languageUser, profileLink, eventName, eventDate string, members string) error {
	email := emailInfo{}
	email.Top = emailTop{Title: Translate("summary_subject", languageUser), To: to}
	email.Body = emailSummaryInfo{
		Subject:        Translate("summary_subject", languageUser),
		Greetings:      Translate("summary_greetings", languageUser),
		Inscriptions:   Translate("summary_inscriptions", languageUser),
		EventFormatted: eventName + " " + Translate("reminder_on_the", languageUser) + " " + eventDate + ".",
		FirstName:      Translate("summary_first_name", languageUser),
		Name:           Translate("summary_name", languageUser),
		Roles:          Translate("summary_roles", languageUser),
		Answer:         Translate("summary_answer", languageUser),
		ParticipateYes: Translate("summary_participate_yes", languageUser),
		ParticipateNo:  Translate("summary_participate_no", languageUser),
		NoAnswer:       Translate("summary_no_answer", languageUser),
		MemberName:     memberName,
		ImageSource:    GetConfigString("domain") + "/static/img/",
		// Members              []model.Member // TO FIX
	}
	email.Bottom = emailBottom{ProfileLink: profileLink, MyProfile: Translate("email_my_profile", languageUser), Suggestions: Translate("email_suggestions", languageUser)}

	emailBodyString, err := email.buildEmail()
	if err != nil {
		return err
	}
	emailString := emailBodyString
	// Send mail
	if err = sendMail([]string{to}, emailString); err != nil {
		Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

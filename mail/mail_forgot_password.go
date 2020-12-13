package mail

import (
	"bytes"
	"html/template"
)

type emailForgotInfo struct {
	Subject                string
	MemberName             string
	SubjectInfo            string
	ResetRequestedTitle    string
	ResetRequestedText     string
	ResetNotRequestedTitle string
	ResetNotRequestedText  string
	Reset                  string
	ResetText              string
	ResetLink              string
	ResetButton            string
	ImageSource            string
	Language               string
}

func (e emailForgotInfo) GetBody() (string, error) {
	t, err := template.ParseFiles("templates/email_forgot_body.html")
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

func SendForgotPasswordEmail(to, memberName, languageUser, resetLink, profileLink string) error {
	email := emailInfo{}
	email.Top = emailTop{Title: Translate("forgot_title", languageUser), To: to}
	email.Body = emailForgotInfo{
		Subject:                Translate("forgot_title", languageUser),
		MemberName:             memberName,
		SubjectInfo:            Translate("forgot_subject_info", languageUser),
		ResetRequestedTitle:    Translate("forgot_reset_requested_title", languageUser),
		ResetRequestedText:     Translate("forgot_reset_requested_text", languageUser),
		ResetNotRequestedTitle: Translate("forgot_reset_not_requested_title", languageUser),
		ResetNotRequestedText:  Translate("forgot_reset_not_requested_text", languageUser),
		Reset:                  Translate("forgot_reset", languageUser),
		ResetText:              Translate("forgot_reset_text", languageUser),
		ResetButton:            Translate("forgot_reset", languageUser),
		ImageSource:            GetConfigString("domain") + "/static/img/",
		ResetLink:              resetLink, Language: languageUser}
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

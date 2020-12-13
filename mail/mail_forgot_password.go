package mail

import (
	"bytes"
	"html/template"

	"github.com/vilisseranen/castellers/common"
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

func SendForgotPasswordEmail(to, memberName, languageUser, resetLink, profileLink string) error {
	email := emailInfo{}
	email.Top = emailTop{Title: common.Translate("forgot_title", languageUser), To: to}
	email.Body = emailForgotInfo{
		Subject:                common.Translate("forgot_title", languageUser),
		MemberName:             memberName,
		SubjectInfo:            common.Translate("forgot_subject_info", languageUser),
		ResetRequestedTitle:    common.Translate("forgot_reset_requested_title", languageUser),
		ResetRequestedText:     common.Translate("forgot_reset_requested_text", languageUser),
		ResetNotRequestedTitle: common.Translate("forgot_reset_not_requested_title", languageUser),
		ResetNotRequestedText:  common.Translate("forgot_reset_not_requested_text", languageUser),
		Reset:                  common.Translate("forgot_reset", languageUser),
		ResetText:              common.Translate("forgot_reset_text", languageUser),
		ResetButton:            common.Translate("forgot_reset", languageUser),
		ImageSource:            common.GetConfigString("domain") + "/static/img/",
		ResetLink:              resetLink, Language: languageUser}
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

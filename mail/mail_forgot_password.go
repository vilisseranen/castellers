package mail

import (
	"bytes"
	"html/template"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

type EmailForgotPasswordPayload struct {
	Member      model.Member      `json:"member"`
	Token       string            `json:"token"`
	Credentials model.Credentials `json:"credentials"`
}

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
	t, err := template.ParseFiles("mail/templates/email_forgot_body.html")
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

func SendForgotPasswordEmail(payload EmailForgotPasswordPayload) error {
	resetLink := common.GetConfigString("domain") + "/reset?" +
		"t=" + payload.Token +
		"&a=reset&u=" + payload.Credentials.Username
	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID
	email := emailInfo{}
	email.Header = emailHeader{
		Title: common.Translate("forgot_title", payload.Member.Language),
	}
	email.Top = emailTop{
		Title:    common.Translate("forgot_title", payload.Member.Language),
		To:       payload.Member.Email,
		Subtitle: common.Translate("forgot_subject_info", payload.Member.Language),
	}
	email.MainSections = []emailMain{
		{
			Title: common.Translate("forgot_reset_not_requested_text", payload.Member.Language),
			Text:  common.Translate("forgot_reset_not_requested_text", payload.Member.Language),
		},
		{
			Title: common.Translate("forgot_reset_requested_title", payload.Member.Language),
			Text:  common.Translate("forgot_reset_requested_text", payload.Member.Language),
		},
	}
	email.Action = emailAction{
		Title: common.Translate("forgot_reset", payload.Member.Language),
		Text:  common.Translate("forgot_reset_text", payload.Member.Language),
		Buttons: []Button{{
			Text: common.Translate("forgot_reset", payload.Member.Language),
			Link: resetLink},
		},
	}
	email.Bottom = emailBottom{
		ProfileLink: profileLink,
		MyProfile:   common.Translate("email_my_profile", payload.Member.Language),
		Suggestions: common.Translate("email_suggestions", payload.Member.Language),
	}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"

	if err := sendMail(email); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

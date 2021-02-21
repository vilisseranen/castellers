package mail

import (
	"bytes"
	"html/template"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

type EmailRegisterPayload struct {
	Member model.Member `json:"member"`
	Author model.Member `json:"author"`
	Token  string       `json:"token"`
}

type emailRegisterInfo struct {
	MemberName            string
	AdminName, AdminExtra string
	LoginLink             string
	ImageSource           string
	Welcome               string
	WelcomeText           string
	NewRegistration       string
	NewRegistrationText   string
	Instructions          string
	Thanks                string
	Confirmation          string
	ConfirmationText      string
	Activation            string
}

func (e emailRegisterInfo) GetBody() (string, error) {
	t, err := template.ParseFiles("mail/templates/email_register_body.html")
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

func SendRegistrationEmail(payload EmailRegisterPayload) error {
	loginLink := common.GetConfigString("domain") + "/reset?" +
		"t=" + payload.Token +
		"&a=activation"
	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID

	email := emailInfo{}
	email.Top = emailTop{Title: common.Translate("registration_title", payload.Member.Language), To: payload.Member.Email}
	email.Body = emailRegisterInfo{
		MemberName: payload.Member.FirstName,
		AdminName:  payload.Author.FirstName, AdminExtra: payload.Author.Extra,
		LoginLink:           loginLink,
		ImageSource:         common.GetConfigString("cdn") + "/static/img/",
		Welcome:             common.Translate("registration_welcome", payload.Member.Language),
		WelcomeText:         common.Translate("registration_welcome_text", payload.Member.Language),
		NewRegistration:     common.Translate("registration_new_title", payload.Member.Language),
		NewRegistrationText: common.Translate("registration_new_text", payload.Member.Language),
		Instructions:        common.Translate("registration_instructions", payload.Member.Language),
		Thanks:              common.Translate("registration_thanks", payload.Member.Language),
		Confirmation:        common.Translate("registration_confirmation_title", payload.Member.Language),
		ConfirmationText:    common.Translate("registration_confirmation_text", payload.Member.Language),
		Activation:          common.Translate("registration_activation", payload.Member.Language),
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

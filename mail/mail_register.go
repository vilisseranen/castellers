package mail

import (
	"bytes"
	"html/template"

	"github.com/vilisseranen/castellers/common"
)

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
	t, err := template.ParseFiles("templates/email_register_body.html")
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

func SendRegistrationEmail(to, memberName, languageUser, adminName, adminExtra, activateLink, profileLink string) error {
	email := emailInfo{}
	email.Top = emailTop{Title: common.Translate("registration_title", languageUser), To: to}
	email.Body = emailRegisterInfo{
		MemberName: memberName,
		AdminName:  adminName, AdminExtra: adminExtra,
		LoginLink:           activateLink,
		ImageSource:         common.GetConfigString("domain") + "/static/img/",
		Welcome:             common.Translate("registration_welcome", languageUser),
		WelcomeText:         common.Translate("registration_welcome_text", languageUser),
		NewRegistration:     common.Translate("registration_new_title", languageUser),
		NewRegistrationText: common.Translate("registration_new_text", languageUser),
		Instructions:        common.Translate("registration_instructions", languageUser),
		Thanks:              common.Translate("registration_thanks", languageUser),
		Confirmation:        common.Translate("registration_confirmation_title", languageUser),
		ConfirmationText:    common.Translate("registration_confirmation_text", languageUser),
		Activation:          common.Translate("registration_activation", languageUser),
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

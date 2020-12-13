package mail

import (
	"bytes"
	"html/template"
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

func SendRegistrationEmail(to, memberName, languageUser, adminName, adminExtra, activateLink, profileLink string) error {
	email := emailInfo{}
	email.Top = emailTop{Title: Translate("registration_title", languageUser), To: to}
	email.Body = emailRegisterInfo{
		MemberName: memberName,
		AdminName:  adminName, AdminExtra: adminExtra,
		LoginLink:           activateLink,
		ImageSource:         GetConfigString("domain") + "/static/img/",
		Welcome:             Translate("registration_welcome", languageUser),
		WelcomeText:         Translate("registration_welcome_text", languageUser),
		NewRegistration:     Translate("registration_new_title", languageUser),
		NewRegistrationText: Translate("registration_new_text", languageUser),
		Instructions:        Translate("registration_instructions", languageUser),
		Thanks:              Translate("registration_thanks", languageUser),
		Confirmation:        Translate("registration_confirmation_title", languageUser),
		ConfirmationText:    Translate("registration_confirmation_text", languageUser),
		Activation:          Translate("registration_activation", languageUser),
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

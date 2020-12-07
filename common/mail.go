package common

import (
	"bytes"
	"html/template"
	"net/smtp"
)

type emailRegisterInfo struct {
	MemberName            string
	Language              string
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

type emailReminderInfo struct {
	MemberName            string
	Language              string
	ParticipationLink     string
	ImageSource           string
	Answer, Participation string
	EventName, EventDate  string
}

type emailSummaryInfo struct {
	MemberName  string
	Language    string
	ImageSource string
	// Members              []model.Member // TO FIX
	Members              string
	EventName, EventDate string
}

type emailTop struct {
	Title string
}

type emailBottom struct {
	Language    string
	ImageSource string
	ProfileLink string
}

func SendRegistrationEmail(to, memberName, languageUser, adminName, adminExtra, activateLink, profileLink string) error {
	title_translated := Translate("registration_title", languageUser)
	header := buildHeader(title_translated, to)
	// Build top of the email
	top := new(bytes.Buffer)
	if err := buildEmailTop(top, emailTop{title_translated}); err != nil {
		Error("Error parsing template: " + err.Error())
		return err
	}
	// Build body
	t, err := template.ParseFiles("templates/email_register_body.html")
	if err != nil {
		Error("Error parsing template: " + err.Error())
		return err
	}
	body := new(bytes.Buffer)
	imageSource := GetConfigString("domain") + "/static/img/"
	emailInfo := emailRegisterInfo{
		memberName, languageUser, adminName, adminExtra, activateLink, imageSource,
		Translate("registration_welcome", languageUser),
		Translate("registration_welcome_text", languageUser),
		Translate("registration_new_title", languageUser),
		Translate("registration_new_text", languageUser),
		Translate("registration_instructions", languageUser),
		Translate("registration_thanks", languageUser),
		Translate("registration_confirmation_title", languageUser),
		Translate("registration_confirmation_text", languageUser),
		Translate("registration_activation", languageUser),
	}
	if err = t.Execute(body, emailInfo); err != nil {
		Error("Error generating template: " + err.Error())
		return err
	}
	// Build bottom of the email
	bottom := new(bytes.Buffer)
	if err := buildEmailBottom(bottom, emailBottom{languageUser, imageSource, profileLink}); err != nil {
		Error("Error parsing template: " + err.Error())
		return err
	}
	email := header + top.String() + body.String() + bottom.String()
	// Send mail
	if err = sendMail([]string{to}, email); err != nil {
		Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

func SendReminderEmail(to, memberName, language, participationLink, profileLink, answer, participation, eventName, eventDate string) error {
	// Prepare header
	var title_translated string
	switch language {
	case "fr":
		title_translated = "Rappel"
	case "en":
		title_translated = "Reminder"
	case "cat":
		title_translated = "Recordatori"
	}
	header := buildHeader(title_translated, to)
	// Build top of the email
	top := new(bytes.Buffer)
	if err := buildEmailTop(top, emailTop{title_translated}); err != nil {
		Error("Error parsing template: " + err.Error())
		return err
	}
	// Build body
	t, err := template.ParseFiles("templates/email_reminder_body.html")
	if err != nil {
		Error("Error parsing template: " + err.Error())
		return err
	}
	body := new(bytes.Buffer)
	imageSource := GetConfigString("domain") + "/static/img/"
	emailInfo := emailReminderInfo{memberName, language, participationLink, imageSource, answer, participation, eventName, eventDate}
	if err = t.Execute(body, emailInfo); err != nil {
		Error("Error generating template: " + err.Error())
		return err
	}
	// Build bottom of the email
	bottom := new(bytes.Buffer)
	if err := buildEmailBottom(bottom, emailBottom{language, imageSource, profileLink}); err != nil {
		Error("Error parsing template: " + err.Error())
		return err
	}
	email := header + top.String() + body.String() + bottom.String()
	// Send mail
	if err = sendMail([]string{to}, email); err != nil {
		Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

// TO FIX
func SendSummaryEmail(to, memberName, language, profileLink, eventName, eventDate string, members string) error {
	// Prepare header
	var title_translated string
	switch language {
	case "fr":
		title_translated = "Inscriptions pour le prochain évènement"
	case "en":
		title_translated = "Inscriptions to the next event"
	case "cat":
		title_translated = "Inscriptions to the next event"
	}
	header := buildHeader(title_translated, to)
	// Build top of the email
	top := new(bytes.Buffer)
	if err := buildEmailTop(top, emailTop{title_translated}); err != nil {
		Error("Error parsing template: " + err.Error())
		return err
	}
	// Build body
	t, err := template.ParseFiles("templates/email_summary_body.html")
	if err != nil {
		Error("Error parsing template: " + err.Error())
		return err
	}
	body := new(bytes.Buffer)
	imageSource := GetConfigString("domain") + "/static/img/"
	emailInfo := emailSummaryInfo{memberName, language, imageSource, members, eventName, eventDate}
	if err = t.Execute(body, emailInfo); err != nil {
		Error("Error generating template: " + err.Error())
		return err
	}
	// Build bottom of the email
	bottom := new(bytes.Buffer)
	if err := buildEmailBottom(bottom, emailBottom{language, imageSource, profileLink}); err != nil {
		Error("Error parsing template: " + err.Error())
		return err
	}
	email := header + top.String() + body.String() + bottom.String()
	// Send mail
	if err = sendMail([]string{to}, email); err != nil {
		Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

func buildHeader(title, to string) string {
	return "Subject: " + title + "\r\n" +
		"To: " + to + "\r\n" +
		"From: Castellers de Montréal <" + GetConfigString("smtp_username") + ">\r\n" +
		"Reply-To: " + GetConfigString("reply_to") + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
		"\r\n"
}

func buildEmailTop(buffer *bytes.Buffer, content emailTop) error {
	t, err := template.ParseFiles("templates/email_top.html")
	if err != nil {
		Error("Error parsing template: " + err.Error())
		return err
	}
	if err = t.Execute(buffer, content); err != nil {
		Error("Error generating template: " + err.Error())
		return err
	}
	return nil
}

func buildEmailBottom(buffer *bytes.Buffer, content emailBottom) error {
	t, err := template.ParseFiles("templates/email_bottom.html")
	if err != nil {
		Error("Error parsing template: " + err.Error())
		return err
	}
	if err = t.Execute(buffer, content); err != nil {
		Error("Error generating template: " + err.Error())
		return err
	}
	return nil
}

func sendMail(to []string, body string) error {
	var auth smtp.Auth
	auth = smtp.PlainAuth("", GetConfigString("smtp_username"), GetConfigString("smtp_password"), GetConfigString("smtp_server"))
	addr := GetConfigString("smtp_server") + ":" + GetConfigString("smtp_port")
	if err := smtp.SendMail(addr, auth, GetConfigString("smtp_username"), to, []byte(body)); err != nil {
		Error(err.Error())
		return err
	}
	return nil
}

package common

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
)

type emailRegisterInfo struct {
	MemberName            string
	Language              string
	AdminName, AdminExtra string
	LoginLink             string
	ImageSource           string
}

type emailReminderInfo struct {
	MemberName            string
	Language              string
	ParticipationLink     string
	ImageSource           string
	Answer, Participation string
	EventName, EventDate  string
}

type emailTop struct {
	Title string
}

type emailBottom struct {
	Language    string
	ImageSource string
	ProfileLink string
}

func SendRegistrationEmail(to, memberName, language, adminName, adminExtra, activateLink, profileLink string) error {
	// Prepare header
	var title_translated string
	switch language {
	case "fr":
		title_translated = "Inscription"
	case "en":
		title_translated = "Inscription"
	case "cat":
		title_translated = "Inscriptió"
	}
	header := buildHeader(title_translated, to)
	// Build top of the email
	top := new(bytes.Buffer)
	if err := buildEmailTop(top, emailTop{title_translated}); err != nil {
		fmt.Println("Error parsing template: " + err.Error())
		return err
	}
	// Build body
	t, err := template.ParseFiles("templates/email_register_body.html")
	if err != nil {
		fmt.Println("Error parsing template: " + err.Error())
		return err
	}
	body := new(bytes.Buffer)
	imageSource := GetConfigString("domain") + "/static/img/"
	emailInfo := emailRegisterInfo{memberName, language, adminName, adminExtra, activateLink, imageSource}
	if err = t.Execute(body, emailInfo); err != nil {
		fmt.Println("Error generating template: " + err.Error())
		return err
	}
	// Build bottom of the email
	bottom := new(bytes.Buffer)
	if err := buildEmailBottom(bottom, emailBottom{language, imageSource, profileLink}); err != nil {
		fmt.Println("Error parsing template: " + err.Error())
		return err
	}
	email := header + top.String() + body.String() + bottom.String()
	// Send mail
	if err = sendMail([]string{to}, email); err != nil {
		fmt.Println("Error sending Email: " + err.Error())
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
		fmt.Println("Error parsing template: " + err.Error())
		return err
	}
	// Build body
	t, err := template.ParseFiles("templates/email_reminder_body.html")
	if err != nil {
		fmt.Println("Error parsing template: " + err.Error())
		return err
	}
	body := new(bytes.Buffer)
	imageSource := GetConfigString("domain") + "/static/img/"
	emailInfo := emailReminderInfo{memberName, language, participationLink, imageSource, answer, participation, eventName, eventDate}
	if err = t.Execute(body, emailInfo); err != nil {
		fmt.Println("Error generating template: " + err.Error())
		return err
	}
	// Build bottom of the email
	bottom := new(bytes.Buffer)
	if err := buildEmailBottom(bottom, emailBottom{language, imageSource, profileLink}); err != nil {
		fmt.Println("Error parsing template: " + err.Error())
		return err
	}
	email := header + top.String() + body.String() + bottom.String()
	// Send mail
	if err = sendMail([]string{to}, email); err != nil {
		fmt.Println("Error sending Email: " + err.Error())
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
		fmt.Println("Error parsing template: " + err.Error())
		return err
	}
	if err = t.Execute(buffer, content); err != nil {
		fmt.Println("Error generating template: " + err.Error())
		return err
	}
	return nil
}

func buildEmailBottom(buffer *bytes.Buffer, content emailBottom) error {
	t, err := template.ParseFiles("templates/email_bottom.html")
	if err != nil {
		fmt.Println("Error parsing template: " + err.Error())
		return err
	}
	if err = t.Execute(buffer, content); err != nil {
		fmt.Println("Error generating template: " + err.Error())
		return err
	}
	return nil
}

func sendMail(to []string, body string) error {
	var auth smtp.Auth
	auth = smtp.PlainAuth("", GetConfigString("smtp_username"), GetConfigString("smtp_password"), GetConfigString("smtp_server"))
	addr := GetConfigString("smtp_server") + ":" + GetConfigString("smtp_port")
	if err := smtp.SendMail(addr, auth, GetConfigString("smtp_username"), to, []byte(body)); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

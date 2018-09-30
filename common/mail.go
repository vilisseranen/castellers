package common

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
)

func SendRegistrationEmail(to, memberName, adminName, adminExtra, activateLink, profileLink string) error {
	// Prepare header
	header := "Subject: Inscription\r\n" +
		"To: " + to + "\r\n" +
		"From: " + GetConfigString("mail_from") + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
		"\r\n"
	// Parse body
	t, err := template.ParseFiles("templates/email_register_fr.html")
	if err != nil {
		fmt.Println(err)
		return err
	}
	buf := new(bytes.Buffer)
	imageSource := GetConfigString("domain") + "/static/img/"
	emailInfo := emailRegisterInfo{memberName, adminName, adminExtra, activateLink, profileLink, imageSource}
	if err = t.Execute(buf, emailInfo); err != nil {
		fmt.Println(err)
		return err
	}
	body := header + buf.String()
	// Send mail
	if err = sendMail([]string{to}, body); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

type emailRegisterInfo struct {
	MemberName             string
	AdminName, AdminExtra  string
	LoginLink, ProfileLink string
	ImageSource            string
}

type email struct {
	from    string
	to      []string
	subject string
	body    string
}

func sendMail(to []string, body string) error {
	var auth smtp.Auth
	auth = smtp.PlainAuth("", GetConfigString("smtp_username"), GetConfigString("smtp_password"), GetConfigString("smtp_server"))
	addr := GetConfigString("smtp_server") + ":" + GetConfigString("smtp_port")
	if err := smtp.SendMail(addr, auth, GetConfigString("mail_from"), to, []byte(body)); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

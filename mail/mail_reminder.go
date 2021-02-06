package mail

import (
	"bytes"
	"html/template"

	"github.com/vilisseranen/castellers/common"
)

type emailReminderInfo struct {
	Subject               string
	Greetings             string
	AnswerNoText          string
	AnswerYesText         string
	PleaseAnswer          string
	CurrentAnswer         string
	AnswerNo              string
	AnswerYes             string
	AnswerChanged         string
	Availability          string
	Confirm               string
	MemberName            string
	ParticipationLink     string
	ImageSource           string
	Answer, Participation string
	EventFormatted        string
}

func (e emailReminderInfo) GetBody() (string, error) {
	t, err := template.ParseFiles("mail/templates/email_reminder_body.html")
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

func SendReminderEmail(to, memberName, languageUser, participationLink, profileLink, answer, participation, eventName, eventDate string) error {
	email := emailInfo{}
	email.Top = emailTop{Title: common.Translate("reminder_subject", languageUser), To: to}
	email.Body = emailReminderInfo{
		Subject:           common.Translate("reminder_subject", languageUser),
		Greetings:         common.Translate("reminder_greetings", languageUser),
		AnswerNoText:      common.Translate("reminder_text_answered_no", languageUser),
		AnswerYesText:     common.Translate("reminder_text_answered_yes", languageUser),
		PleaseAnswer:      common.Translate("reminder_please_answer", languageUser),
		CurrentAnswer:     common.Translate("reminder_current_answer", languageUser),
		AnswerNo:          common.Translate("reminder_answer_no", languageUser),
		AnswerYes:         common.Translate("reminder_answer_yes", languageUser),
		AnswerChanged:     common.Translate("reminder_answer_changed", languageUser),
		Availability:      common.Translate("reminder_availability", languageUser),
		Confirm:           common.Translate("reminder_confirm", languageUser),
		MemberName:        memberName,
		ParticipationLink: participationLink,
		ImageSource:       common.GetConfigString("domain") + "/static/img/",
		Answer:            answer, Participation: participation,
		EventFormatted: eventName + " " + common.Translate("reminder_on_the", languageUser) + " " + eventDate + ".",
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

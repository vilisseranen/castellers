package mail

import (
	"bytes"
	"html/template"
	"time"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

type EmailReminderPayload struct {
	Member        model.Member        `json:"member"`
	Event         model.Event         `json:"event"`
	Participation model.Participation `json:"participation"`
	Token         string              `json:"token"`
}

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

func SendReminderEmail(payload EmailReminderPayload) error {
	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID
	participationLink := common.GetConfigString("domain") + "/events?" +
		"a=participate" +
		"&e=" + payload.Event.UUID +
		"&u=" + payload.Member.UUID +
		"&t=" + payload.Token +
		"&p="
	answer := "false"
	if payload.Participation.Answer == common.AnswerYes || payload.Participation.Answer == common.AnswerNo {
		answer = "true"
	}
	location, err := time.LoadLocation("America/Montreal")
	if err != nil {
		common.Error("%v\n", err)
		return err
	}
	eventDate := time.Unix(int64(payload.Event.StartDate), 0).In(location).Format("02-01-2006")
	email := emailInfo{}
	email.Top = emailTop{Title: common.Translate("reminder_subject", payload.Member.Language), To: payload.Member.Email}
	email.Body = emailReminderInfo{
		Subject:           common.Translate("reminder_subject", payload.Member.Language),
		Greetings:         common.Translate("reminder_greetings", payload.Member.Language),
		AnswerNoText:      common.Translate("reminder_text_answered_no", payload.Member.Language),
		AnswerYesText:     common.Translate("reminder_text_answered_yes", payload.Member.Language),
		PleaseAnswer:      common.Translate("reminder_please_answer", payload.Member.Language),
		CurrentAnswer:     common.Translate("reminder_current_answer", payload.Member.Language),
		AnswerNo:          common.Translate("reminder_answer_no", payload.Member.Language),
		AnswerYes:         common.Translate("reminder_answer_yes", payload.Member.Language),
		AnswerChanged:     common.Translate("reminder_answer_changed", payload.Member.Language),
		Availability:      common.Translate("reminder_availability", payload.Member.Language),
		Confirm:           common.Translate("reminder_confirm", payload.Member.Language),
		MemberName:        payload.Member.FirstName,
		ParticipationLink: participationLink,
		ImageSource:       common.GetConfigString("cdn") + "/static/img/",
		Answer:            answer, Participation: payload.Participation.Answer,
		EventFormatted: payload.Event.Name + " " + common.Translate("reminder_on_the", payload.Member.Language) + " " + eventDate + ".",
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

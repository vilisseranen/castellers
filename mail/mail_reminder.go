package mail

import (
	"context"
	"time"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
	"go.elastic.co/apm"
)

type EmailReminderPayload struct {
	Member        model.Member        `json:"member"`
	Event         model.Event         `json:"event"`
	Participation model.Participation `json:"participation"`
	Token         string              `json:"token"`
}

func SendReminderEmail(ctx context.Context, payload EmailReminderPayload) error {
	span, ctx := apm.StartSpan(ctx, "mail.SendReminderEmail", APM_SPAN_TYPE_CRON)
	defer span.End()

	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID
	participationLink := common.GetConfigString("domain") + "/eventShow/" + payload.Event.UUID + "?" +
		"a=participate" +
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
	email.Header = emailHeader{Title: common.Translate("reminder_subject", payload.Member.Language)}
	email.Top = emailTop{
		Title:    common.Translate("greetings", payload.Member.Language) + " " + payload.Member.FirstName,
		Subtitle: common.Translate("reminder_text_answered_"+answer, payload.Member.Language),
		To:       payload.Member.Email}
	mainSection := emailMain{
		Title: payload.Event.Name + " " + common.Translate("on_the", payload.Member.Language) + " " + eventDate + ".",
	}
	if answer == "true" {
		// The member gave an answer
		mainSection.Subtitle = common.Translate("reminder_current_answer", payload.Member.Language)
		mainSection.Text = common.Translate("reminder_answer_"+payload.Participation.Answer, payload.Member.Language)
	} else {
		mainSection.Text = common.Translate("reminder_please_answer", payload.Member.Language)
	}
	email.MainSections = []emailMain{mainSection}
	email.Action = emailAction{
		Title: common.Translate("reminder_availability", payload.Member.Language),
		Text:  common.Translate("reminder_confirm", payload.Member.Language),
		Buttons: []Button{
			{
				Text:  common.Translate("reminder_answer_yes", payload.Member.Language),
				Link:  participationLink + "yes",
				Color: "#20470b",
			},
			{
				Text:  common.Translate("reminder_answer_no", payload.Member.Language),
				Link:  participationLink + "no",
				Color: "#aa0000",
			},
		},
	}
	email.Bottom = emailBottom{ProfileLink: profileLink, MyProfile: common.Translate("email_my_profile", payload.Member.Language), Suggestions: common.Translate("email_suggestions", payload.Member.Language)}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"

	if err = sendMail(ctx, email); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

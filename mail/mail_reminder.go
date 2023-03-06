package mail

import (
	"context"
	"fmt"
	"time"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

type EmailReminderPayload struct {
	Member        model.Member
	Event         model.Event
	Participation model.Participation
	Token         string
	Dependents    []model.Member
}

// Will send an email reminding members to register for an event
// If the member has dependents, the responsible will receive one
// email per dependent in addition to their own
func SendReminderEmail(ctx context.Context, payload EmailReminderPayload) error {
	ctx, span := tracer.Start(ctx, "mail.SendReminderEmail")
	defer span.End()

	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID
	participationLink := common.GetConfigString("domain") + "/eventShow/" + payload.Event.UUID + "?" +
		"a=participate" +
		"&m=" + payload.Member.UUID +
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
	email.Header = emailHeader{
		Title: payload.Member.FirstName + ", " + common.Translate("reminder_subject", payload.Member.Language) + " " + eventDate,
	}
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
	email.Actions = []emailAction{{
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
	}}

	// If the member has dependents, he will have an addition Action section for each dependent

	for _, dependent := range payload.Dependents {
		participationLink = common.GetConfigString("domain") + "/eventShow/" + payload.Event.UUID + "?" +
			"a=participate" +
			"&m=" + dependent.UUID +
			"&t=" + payload.Token +
			"&p="
		email.Actions = append(email.Actions, emailAction{
			Title: common.Translate("reminder_availability_dependent", payload.Member.Language) + dependent.FirstName,
			Text:  fmt.Sprintf(common.Translate("reminder_confirm_dependent", payload.Member.Language), dependent.FirstName),
			Buttons: []Button{
				{
					Text:  common.Translate("reminder_answer_yes_dependent", payload.Member.Language),
					Link:  participationLink + "yes",
					Color: "#20470b",
				},
				{
					Text:  common.Translate("reminder_answer_no_dependent", payload.Member.Language),
					Link:  participationLink + "no",
					Color: "#aa0000",
				},
			},
		})
	}

	email.Bottom = emailBottom{ProfileLink: profileLink, MyProfile: common.Translate("email_my_profile", payload.Member.Language), Suggestions: common.Translate("email_suggestions", payload.Member.Language)}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"

	if err = sendMail(ctx, email); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}

	return nil
}

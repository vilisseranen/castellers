package mail

import (
	"context"
	"time"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
	"go.elastic.co/apm"
)

type EmailDeletedEventPayload struct {
	Member       model.Member `json:"member"`
	EventDeleted model.Event  `json:"eventDeleted"`
}

func SendDeletedEventEmail(ctx context.Context, payload EmailDeletedEventPayload) error {
	span, ctx := apm.StartSpan(ctx, "mail.SendDeletedEventEmail", APM_SPAN_TYPE_CRON)
	defer span.End()

	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID
	location, err := time.LoadLocation("America/Montreal")
	if err != nil {
		common.Error("%v\n", err)
		return err
	}
	eventDate := time.Unix(int64(payload.EventDeleted.StartDate), 0).In(location).Format("02-01-2006")

	// Build email
	email := emailInfo{}
	email.Header = emailHeader{Title: common.Translate("deleted_event_subject", payload.Member.Language)}
	email.Top = emailTop{
		Title:    common.Translate("greetings", payload.Member.Language) + " " + payload.Member.FirstName,
		Subtitle: common.Translate("deleted_event_intro", payload.Member.Language),
		To:       payload.Member.Email}
	email.MainSections = []emailMain{{
		Title: payload.EventDeleted.Name + " " + common.Translate("on_the", payload.Member.Language) + " " + eventDate + ".",
		Text:  common.Translate("deleted_event_text", payload.Member.Language)}}
	email.Action = emailAction{}
	email.Bottom = emailBottom{ProfileLink: profileLink, MyProfile: common.Translate("email_my_profile", payload.Member.Language), Suggestions: common.Translate("email_suggestions", payload.Member.Language)}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"

	if err = sendMail(ctx, email); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

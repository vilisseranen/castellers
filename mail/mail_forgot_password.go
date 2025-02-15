package mail

import (
	"context"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

type EmailForgotPasswordPayload struct {
	Member      model.Member      `json:"member"`
	Token       string            `json:"token"`
	Credentials model.Credentials `json:"credentials"`
}

func SendForgotPasswordEmail(ctx context.Context, payload EmailForgotPasswordPayload) error {
	ctx, span := tracer.Start(ctx, "mail.SendForgotPasswordEmail")
	defer span.End()

	resetLink := common.GetConfigString("domain") + "/reset?" +
		"t=" + payload.Token
	if payload.Credentials.Username != "" {
		resetLink += "&u=" + payload.Credentials.Username + "&a=reset"
	} else {
		resetLink += "&a=activation"
	}
	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID
	email := emailInfo{}
	email.Header = emailHeader{
		Title: common.Translate("forgot_title", payload.Member.Language),
	}
	email.Top = emailTop{
		Title:    common.Translate("forgot_title", payload.Member.Language),
		To:       payload.Member.Email,
		Subtitle: common.Translate("forgot_subject_info", payload.Member.Language),
	}
	email.MainSections = []emailMain{
		{
			Title: common.Translate("forgot_reset_not_requested_title", payload.Member.Language),
			Text:  common.Translate("forgot_reset_not_requested_text", payload.Member.Language),
		},
		{
			Title: common.Translate("forgot_reset_requested_title", payload.Member.Language),
			Text:  common.Translate("forgot_reset_requested_text", payload.Member.Language),
		},
	}
	email.Actions = []emailAction{{
		Title: common.Translate("forgot_reset", payload.Member.Language),
		Text:  common.Translate("forgot_reset_text", payload.Member.Language),
		Buttons: []Button{{
			Text: common.Translate("forgot_reset", payload.Member.Language),
			Link: resetLink},
		},
	}}
	email.Bottom = emailBottom{
		ProfileLink: profileLink,
		MyProfile:   common.Translate("email_my_profile", payload.Member.Language),
		Suggestions: common.Translate("email_suggestions", payload.Member.Language),
	}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"

	if err := sendMail(ctx, email); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

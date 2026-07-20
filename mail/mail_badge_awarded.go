package mail

import (
	"context"
	"fmt"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

type EmailBadgeAwardedPayload struct {
	Member    model.Member
	BadgeCode string
}

// SendBadgeAwardedEmail congratulates a member for a newly awarded badge
// and links them to their badge board.
func SendBadgeAwardedEmail(ctx context.Context, payload EmailBadgeAwardedPayload) error {
	ctx, span := tracer.Start(ctx, "mail.SendBadgeAwardedEmail")
	defer span.End()

	lang := payload.Member.Language
	badgeName := common.Translate("badge_name_"+payload.BadgeCode, lang)
	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID
	badgesLink := common.GetConfigString("domain") + "/myBadges"

	email := emailInfo{}
	email.Header = emailHeader{
		Title: common.Translate("badge_awarded_subject", lang),
	}
	email.Top = emailTop{
		Title:    common.Translate("greetings", lang) + " " + payload.Member.FirstName,
		Subtitle: common.Translate("badge_awarded_intro", lang),
		To:       payload.Member.Email,
	}
	email.MainSections = []emailMain{{
		Title: common.Translate("badge_awarded_title", lang),
		Text:  fmt.Sprintf(common.Translate("badge_awarded_text", lang), badgeName),
	}}
	email.Actions = []emailAction{{
		Title: common.Translate("badge_awarded_action_title", lang),
		Text:  common.Translate("badge_awarded_action_text", lang),
		Buttons: []Button{{
			Text: common.Translate("badge_awarded_action_button", lang),
			Link: badgesLink,
		}},
	}}
	email.Bottom = emailBottom{
		ProfileLink: profileLink,
		MyProfile:   common.Translate("email_my_profile", lang),
		Suggestions: common.Translate("email_suggestions", lang),
	}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"

	if err := sendMail(ctx, email); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

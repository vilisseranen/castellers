package mail

import (
	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

type EmailRegisterPayload struct {
	Member model.Member `json:"member"`
	Author model.Member `json:"author"`
	Token  string       `json:"token"`
}

func SendRegistrationEmail(payload EmailRegisterPayload) error {
	loginLink := common.GetConfigString("domain") + "/reset?" +
		"t=" + payload.Token +
		"&a=activation"
	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID

	email := emailInfo{}
	email.Header = emailHeader{Title: common.Translate("registration_title", payload.Member.Language)}
	email.Top = emailTop{
		Title:    common.Translate("registration_welcome", payload.Member.Language) + " " + payload.Member.FirstName,
		Subtitle: common.Translate("registration_welcome_text", payload.Member.Language),
		To:       payload.Member.Email}
	email.MainSections = []emailMain{{
		Title:    common.Translate("registration_new_title", payload.Member.Language),
		Subtitle: common.Translate("registration_new_text", payload.Member.Language),
		Text:     common.Translate("registration_instructions", payload.Member.Language) + "<br/><br/>" + common.Translate("registration_thanks", payload.Member.Language),
		Author:   payload.Author.FirstName + " " + payload.Author.LastName,
	}}
	email.Action = emailAction{
		Title: common.Translate("registration_confirmation_title", payload.Member.Language),
		Text:  common.Translate("registration_confirmation_text", payload.Member.Language),
		Buttons: []Button{{
			Text: common.Translate("registration_activation", payload.Member.Language),
			Link: loginLink,
		}},
	}
	email.Bottom = emailBottom{ProfileLink: profileLink, MyProfile: common.Translate("email_my_profile", payload.Member.Language), Suggestions: common.Translate("email_suggestions", payload.Member.Language)}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"

	if err := sendMail(email); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

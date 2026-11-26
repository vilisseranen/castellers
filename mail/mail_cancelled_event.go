package mail

import (
	"context"
	"html/template"
	"strings"
	"time"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

// EmailCancelledEventPayload is the email telling a member that an event is
// cancelled or, when Restored is true, that it takes place after all.
type EmailCancelledEventPayload struct {
	Member   model.Member `json:"member"`
	Event    model.Event  `json:"event"`
	Restored bool         `json:"restored"`
}

// cancelledEventReason writes why an event is cancelled, as HTML: the reason,
// then the note of the admin, if any.
func cancelledEventReason(event model.Event, language string) string {
	reason := ""
	if common.StringInSlice(event.CancelReason, model.ValidCancelReasons) {
		reason = "<strong>" + common.Translate("cancelled_event_reason", language) + "</strong> " +
			common.Translate("cancelled_event_reason_"+event.CancelReason, language)
	}
	note := template.HTMLEscapeString(strings.TrimSpace(event.CancelNote))
	note = strings.ReplaceAll(note, "\n", "<br/>")
	if reason != "" && note != "" {
		return reason + "<br/>" + note
	}
	return reason + note
}

// cancelledEventEmail builds the email of a cancelled, or restored, event.
func cancelledEventEmail(payload EmailCancelledEventPayload, location *time.Location, now time.Time) emailInfo {
	language := payload.Member.Language
	prefix := "cancelled_event_"
	if payload.Restored {
		prefix = "restored_event_"
	}
	text := common.Translate(prefix+"text", language)
	if !payload.Restored {
		if reason := cancelledEventReason(payload.Event, language); reason != "" {
			text += "<br/><br/>" + reason
		}
	}
	email := emailInfo{}
	email.Header = emailHeader{Title: common.Translate(prefix+"subject", language) + " " + payload.Event.Name}
	email.Top = emailTop{
		Title:    common.Translate("greetings", language) + " " + payload.Member.FirstName,
		Subtitle: common.Translate(prefix+"intro", language),
		To:       payload.Member.Email}
	email.MainSections = []emailMain{{
		Title: payload.Event.Name,
		Text:  formatWhen(payload.Event, location, now, language) + "<br/><br/>" + text}}
	email.Actions = []emailAction{{}}
	if payload.Restored {
		email.Actions = []emailAction{{
			Title: common.Translate("modified_event_action_title", language),
			Text:  common.Translate("modified_event_action_text", language),
			Buttons: []Button{{
				Text: common.Translate("modified_event_action_button", language),
				Link: common.GetConfigString("domain") + "/eventShow/" + payload.Event.UUID,
			}},
		}}
	}
	email.Bottom = emailBottom{
		ProfileLink: common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID,
		MyProfile:   common.Translate("email_my_profile", language),
		Suggestions: common.Translate("email_suggestions", language)}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"
	return email
}

func SendCancelledEventEmail(ctx context.Context, payload EmailCancelledEventPayload) error {
	ctx, span := tracer.Start(ctx, "mail.SendCancelledEventEmail")
	defer span.End()

	location, err := time.LoadLocation("America/Montreal")
	if err != nil {
		common.Error("%v\n", err)
		return err
	}
	if err = sendMail(ctx, cancelledEventEmail(payload, location, time.Now())); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

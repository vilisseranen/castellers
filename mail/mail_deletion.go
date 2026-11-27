package mail

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

// EmailDeletionNoticePayload announces the automatic deletion of Member to
// Recipient: the member themselves, or somebody responsible for them.
type EmailDeletionNoticePayload struct {
	Recipient      model.Member
	Member         model.Member
	DeletionAt     int64 // Unix seconds
	NeverActivated bool  // the account was created and never used
}

// EmailDeletionDigestPayload tells an admin about the planned deletions and
// the coming purges.
type EmailDeletionDigestPayload struct {
	Member    model.Member
	Deletions []model.AdminNotice
	Purges    []model.AdminNotice
}

// deletionDay writes a day as in the other emails: 31-12-2026, in Montréal.
func deletionDay(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	if location, err := time.LoadLocation("America/Montreal"); err == nil {
		t = t.In(location)
	}
	return t.Format("02-01-2006")
}

// months turns a delay in days into months, for the texts: 730 => 24.
func months(days int) int {
	return days / 30
}

// SendDeletionNoticeEmail tells a member that their account will be deleted
// because they have been inactive for too long, and how to keep it.
func SendDeletionNoticeEmail(ctx context.Context, payload EmailDeletionNoticePayload) error {
	ctx, span := tracer.Start(ctx, "mail.SendDeletionNoticeEmail")
	defer span.End()

	email := deletionNoticeEmail(payload,
		months(common.GetConfigInt("inactive_delete_delay_days")),
		months(common.GetConfigInt("pending_delete_delay_days")),
		common.GetConfigInt("purge_delay_days"))
	if err := sendMail(ctx, email); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

// deletionNoticeEmail writes the notice. inactiveMonths and pendingMonths are
// the delays after which a member in pause and an account never activated are
// deleted; purgeDays is the time before the data of a deleted member is erased.
func deletionNoticeEmail(payload EmailDeletionNoticePayload, inactiveMonths, pendingMonths, purgeDays int) emailInfo {
	lang := payload.Recipient.Language
	date := deletionDay(payload.DeletionAt)
	forSelf := payload.Recipient.UUID == payload.Member.UUID

	email := emailInfo{}
	email.Top = emailTop{
		Title:    common.Translate("greetings", lang) + " " + payload.Recipient.FirstName,
		Subtitle: common.Translate("deletion_notice_intro", lang),
		To:       payload.Recipient.Email,
	}
	if forSelf {
		reason := fmt.Sprintf(common.Translate("deletion_notice_text_inactive", lang), inactiveMonths)
		if payload.NeverActivated {
			reason = fmt.Sprintf(common.Translate("deletion_notice_text_pending", lang), pendingMonths)
		}
		title := fmt.Sprintf(common.Translate("deletion_notice_title", lang), date)
		email.Header = emailHeader{Title: title}
		email.MainSections = []emailMain{{
			Title: title,
			Text:  html.EscapeString(reason + " " + fmt.Sprintf(common.Translate("deletion_notice_nothing", lang), purgeDays)),
		}}
		email.Actions = []emailAction{{
			Title: common.Translate("deletion_notice_action_title", lang),
			Text:  common.Translate("deletion_notice_action_text", lang),
			Buttons: []Button{{
				Text: common.Translate("deletion_notice_action_button", lang),
				Link: common.GetConfigString("domain") + "/login",
			}},
		}}
	} else {
		name := payload.Member.FirstName
		title := fmt.Sprintf(common.Translate("deletion_notice_dependent_title", lang), name, date)
		reason := fmt.Sprintf(common.Translate("deletion_notice_dependent_text", lang), name, inactiveMonths)
		email.Header = emailHeader{Title: title}
		email.MainSections = []emailMain{
			{
				Title: title,
				Text:  html.EscapeString(reason + " " + fmt.Sprintf(common.Translate("deletion_notice_dependent_nothing", lang), purgeDays)),
			},
			{
				Title: common.Translate("deletion_notice_dependent_action_title", lang),
				Text:  html.EscapeString(common.Translate("deletion_notice_dependent_action_text", lang)),
			},
		}
	}
	email.Bottom = emailBottom{
		ProfileLink: common.GetConfigString("domain") + "/memberEdit/" + payload.Recipient.UUID,
		MyProfile:   common.Translate("email_my_profile", lang),
		Suggestions: common.Translate("email_suggestions", lang),
	}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"
	return email
}

// noticesList writes a list of members with a day: "Maria Soler — 31-12-2026".
func noticesList(notices []model.AdminNotice) string {
	var list strings.Builder
	list.WriteString("<ul>")
	for _, notice := range notices {
		list.WriteString("<li>" + html.EscapeString(notice.Name) + " — " + deletionDay(notice.Date) + "</li>")
	}
	list.WriteString("</ul>")
	return list.String()
}

// SendDeletionDigestEmail tells an admin which members will be deleted
// automatically and which deleted members will soon be erased for good.
func SendDeletionDigestEmail(ctx context.Context, payload EmailDeletionDigestPayload) error {
	ctx, span := tracer.Start(ctx, "mail.SendDeletionDigestEmail")
	defer span.End()

	if err := sendMail(ctx, deletionDigestEmail(payload)); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

func deletionDigestEmail(payload EmailDeletionDigestPayload) emailInfo {
	lang := payload.Member.Language

	email := emailInfo{}
	email.Header = emailHeader{Title: common.Translate("deletion_digest_subject", lang)}
	email.Top = emailTop{
		Title:    common.Translate("greetings", lang) + " " + payload.Member.FirstName,
		Subtitle: common.Translate("deletion_digest_intro", lang),
		To:       payload.Member.Email,
	}
	if len(payload.Deletions) > 0 {
		email.MainSections = append(email.MainSections, emailMain{
			Title: common.Translate("deletion_digest_deletions_title", lang),
			Text:  html.EscapeString(common.Translate("deletion_digest_deletions_text", lang)) + noticesList(payload.Deletions),
		})
	}
	if len(payload.Purges) > 0 {
		email.MainSections = append(email.MainSections, emailMain{
			Title: common.Translate("deletion_digest_purges_title", lang),
			Text:  html.EscapeString(common.Translate("deletion_digest_purges_text", lang)) + noticesList(payload.Purges),
		})
	}
	email.Actions = []emailAction{{
		Title: common.Translate("deletion_digest_action_title", lang),
		Text:  common.Translate("deletion_digest_action_text", lang),
		Buttons: []Button{{
			Text: common.Translate("deletion_digest_action_button", lang),
			Link: common.GetConfigString("domain") + "/members?upcoming=1",
		}},
	}}
	email.Bottom = emailBottom{
		ProfileLink: common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID,
		MyProfile:   common.Translate("email_my_profile", lang),
		Suggestions: common.Translate("email_suggestions", lang),
	}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"
	return email
}

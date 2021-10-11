package mail

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
	"go.elastic.co/apm"
)

type EmailSummaryPayload struct {
	Member       model.Member   `json:"member"`
	Event        model.Event    `json:"event"`
	Participants []model.Member `json:"participants"`
}

func SendSummaryEmail(ctx context.Context, payload EmailSummaryPayload) error {
	span, ctx := apm.StartSpan(ctx, "mail.SendSummaryEmail", APM_SPAN_TYPE_CRON)
	defer span.End()

	common.Debug("Send summary Event Email")
	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID
	var location, err = time.LoadLocation("America/Montreal")
	if err != nil {
		common.Error("%v\n", err)
		return err
	}
	eventDate := time.Unix(int64(payload.Event.StartDate), 0).In(location).Format("02-01-2006")
	summary := summary{
		FirstName:      common.Translate("summary_first_name", payload.Member.Language),
		Name:           common.Translate("summary_name", payload.Member.Language),
		Roles:          common.Translate("summary_roles", payload.Member.Language),
		Answer:         common.Translate("summary_answer", payload.Member.Language),
		ParticipateYes: common.Translate("summary_participate_yes", payload.Member.Language),
		ParticipateNo:  common.Translate("summary_participate_no", payload.Member.Language),
		NoAnswer:       common.Translate("summary_no_answer", payload.Member.Language),
		Members:        payload.Participants,
	}
	email := emailInfo{}
	email.Header = emailHeader{common.Translate("summary_subject", payload.Member.Language)}
	email.Top = emailTop{
		Title:    common.Translate("greetings", payload.Member.Language) + " " + payload.Member.FirstName,
		Subtitle: common.Translate("summary_inscriptions", payload.Member.Language),
		To:       payload.Member.Email,
	}
	summaryTable, err := summaryTable(summary)
	if err != nil {
		return err
	}
	// count number of castellers registered for the event
	registeredForEvent := 0
	for _, m := range summary.Members {
		if m.Participation == "yes" {
			registeredForEvent += 1
		}
	}
	email.MainSections = []emailMain{{
		Title:    payload.Event.Name + " " + common.Translate("on_the", payload.Member.Language) + " " + eventDate + ".",
		Subtitle: fmt.Sprintf(common.Translate("registeredForEvent", payload.Member.Language), registeredForEvent),
		Text:     summaryTable,
	}}
	email.Bottom = emailBottom{ProfileLink: profileLink, MyProfile: common.Translate("email_my_profile", payload.Member.Language), Suggestions: common.Translate("email_suggestions", payload.Member.Language)}
	email.ImageSource = common.GetConfigString("cdn") + "/static/img/"

	if err = sendMail(ctx, email); err != nil {
		common.Error("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

type summary struct {
	FirstName      string
	Name           string
	Roles          string
	Answer         string
	ParticipateYes string
	ParticipateNo  string
	NoAnswer       string
	Members        []model.Member
}

func summaryTable(summary summary) (string, error) {
	const templateChanges = `              <table class="pure-table pure-table-bordered" style="border: 1px solid #ccc; margin: 50px auto;">
	<thead style="background: #3498db; color: white; font-weight:bold;">
	  <tr>
		<th>{{ .FirstName }}</th>
		<th>{{ .Name }}</th>
		<th>{{ .Roles }}</th>
		<th>{{ .Answer }}</th>
	  </tr>
	</thead>
	<tbody>
	  {{ range .Members }}
	  <tr>
		<td>{{ .FirstName }}</td>
		<td>{{ .LastName }}</td>
		<td>{{ .Roles }}</td>
		<td>
		  {{ if eq .Participation "yes" }}
		  {{ $.ParticipateYes }}
		  {{ else if eq .Participation "no" }}
		  {{ $.ParticipateNo }}
		  {{ else }}
		  {{ $.NoAnswer }}
		  {{ end }}
		</td>
	  </tr>
	  {{ end }}
	</tbody>
  </table>`
	t, err := template.New("summary").Parse(templateChanges)
	if err != nil {
		common.Error("Error parsing template: " + err.Error())
		return "", err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, summary); err != nil {
		common.Error("Error generating template: " + err.Error())
		return "", err
	}
	return buffer.String(), nil
}

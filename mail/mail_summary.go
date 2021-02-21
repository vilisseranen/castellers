package mail

import (
	"bytes"
	"html/template"
	"time"

	"github.com/vilisseranen/castellers/common"
	"github.com/vilisseranen/castellers/model"
)

type EmailSummaryPayload struct {
	Member       model.Member   `json:"member"`
	Event        model.Event    `json:"event"`
	Participants []model.Member `json:"participants"`
}
type emailSummaryInfo struct {
	Subject        string
	Greetings      string
	Inscriptions   string
	EventFormatted string
	FirstName      string
	Name           string
	Roles          string
	Answer         string
	ParticipateYes string
	ParticipateNo  string
	NoAnswer       string
	MemberName     string
	ImageSource    string
	Members        []model.Member
}

func (e emailSummaryInfo) GetBody() (string, error) {
	t, err := template.ParseFiles("mail/templates/email_summary_body.html")
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

func SendSummaryEmail(payload EmailSummaryPayload) error {
	profileLink := common.GetConfigString("domain") + "/memberEdit/" + payload.Member.UUID
	var location, err = time.LoadLocation("America/Montreal")
	if err != nil {
		common.Error("%v\n", err)
		return err
	}
	eventDate := time.Unix(int64(payload.Event.StartDate), 0).In(location).Format("02-01-2006")

	email := emailInfo{}
	email.Top = emailTop{Title: common.Translate("summary_subject", payload.Member.Language), To: payload.Member.Email}
	email.Body = emailSummaryInfo{
		Subject:        common.Translate("summary_subject", payload.Member.Language),
		Greetings:      common.Translate("summary_greetings", payload.Member.Language),
		Inscriptions:   common.Translate("summary_inscriptions", payload.Member.Language),
		EventFormatted: payload.Event.Name + " " + common.Translate("reminder_on_the", payload.Member.Language) + " " + eventDate + ".",
		FirstName:      common.Translate("summary_first_name", payload.Member.Language),
		Name:           common.Translate("summary_name", payload.Member.Language),
		Roles:          common.Translate("summary_roles", payload.Member.Language),
		Answer:         common.Translate("summary_answer", payload.Member.Language),
		ParticipateYes: common.Translate("summary_participate_yes", payload.Member.Language),
		ParticipateNo:  common.Translate("summary_participate_no", payload.Member.Language),
		NoAnswer:       common.Translate("summary_no_answer", payload.Member.Language),
		MemberName:     payload.Member.FirstName,
		ImageSource:    common.GetConfigString("cdn") + "/static/img/",
		Members:        payload.Participants,
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

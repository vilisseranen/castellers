package mail

import (
	"bytes"
	"context"
	"html/template"
	"net/smtp"

	"github.com/vilisseranen/castellers/common"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("castellers")

type emailInfo struct {
	Header       emailHeader
	Top          emailTop
	MainSections []emailMain
	Action       emailAction
	Bottom       emailBottom
	ImageSource  string
}

type emailHeader struct {
	Title string
}

type emailTop struct {
	Title    string
	Subtitle string
	To       string
}

type emailMain struct {
	Title    string
	Subtitle string
	Text     string
	Author   string
}

type emailAction struct {
	Title   string
	Text    string
	Buttons []Button
}

type Button struct {
	Text  string
	Link  string
	Color string
}

type emailBottom struct {
	ProfileLink string
	MyProfile   string
	Suggestions string
}

func unescape(s string) template.HTML {
	return template.HTML(s)
}

func (e emailInfo) buildEmail() (string, error) {
	t, err := template.New("email.html").Funcs(template.FuncMap{"unescape": unescape}).ParseFiles("mail/templates/email.html")
	if err != nil {
		common.Error("Error parsing template: " + err.Error())
		return "", err
	}
	// t.Funcs(template.FuncMap{"unescape": unescape})
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, e); err != nil {
		common.Error("Error generating template: " + err.Error())
		return "", err
	}
	header := buildHeader(e.Header.Title, e.Top.To)
	return header + buffer.String(), nil
}

func buildHeader(title, to string) string {
	return "Subject: " + title + "\r\n" +
		"To: " + to + "\r\n" +
		"From: Castellers de Montréal <" + common.GetConfigString("smtp_username") + ">\r\n" +
		"Reply-To: " + common.GetConfigString("reply_to") + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
		"\r\n"
}

func sendMail(ctx context.Context, email emailInfo) error {
	ctx, span := tracer.Start(ctx, "mail.sendMail")
	defer span.End()

	body, err := email.buildEmail()
	if err != nil {
		common.Error("Cannot build Email")
		return err
	}

	var auth smtp.Auth
	auth = smtp.PlainAuth("", common.GetConfigString("smtp_username"), common.GetConfigString("smtp_password"), common.GetConfigString("smtp_server"))
	addr := common.GetConfigString("smtp_server") + ":" + common.GetConfigString("smtp_port")
	if err := smtp.SendMail(addr, auth, common.GetConfigString("smtp_username"), []string{email.Top.To}, []byte(body)); err != nil {
		common.Error(err.Error())
		return err
	}
	return nil
}

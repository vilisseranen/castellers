package mail

import (
	"bytes"
	"html/template"
	"net/smtp"

	"github.com/vilisseranen/castellers/common"
)

type emailInfo struct {
	Top    emailTop
	Body   emailBody
	Bottom emailBottom
}

type emailTop struct {
	Title string
	To    string
}

type emailBody interface {
	GetBody() (string, error)
}

type emailBottom struct {
	ImageSource string
	ProfileLink string
	MyProfile   string
	Suggestions string
}

func (e emailInfo) buildEmail() (string, error) {
	emailTop, err := e.Top.GetTop()
	if err != nil {
		return "", nil
	}
	emailBody, err := e.Body.GetBody()
	if err != nil {
		return "", nil
	}
	emailBottom, err := e.Bottom.GetBottom()
	if err != nil {
		return "", nil
	}
	return emailTop + emailBody + emailBottom, nil
}

func (e emailTop) GetTop() (string, error) {
	t, err := template.ParseFiles("mail/templates/email_top.html")
	if err != nil {
		common.Error("Error parsing template: " + err.Error())
		return "", err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, e); err != nil {
		common.Error("Error generating template: " + err.Error())
		return "", err
	}
	header := buildHeader(e.Title, e.To)
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

func (e emailBottom) GetBottom() (string, error) {
	e.ImageSource = common.GetConfigString("cdn") + "/static/img/"
	t, err := template.ParseFiles("mail/templates/email_bottom.html")
	if err != nil {
		common.Error("Error parsing template: " + err.Error())
		return "", err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, e); err != nil {
		common.Error("Error generating template: " + err.Error())
		return "", err
	}
	return buffer.String(), nil
}

func sendMail(to []string, body string) error {
	var auth smtp.Auth
	auth = smtp.PlainAuth("", common.GetConfigString("smtp_username"), common.GetConfigString("smtp_password"), common.GetConfigString("smtp_server"))
	addr := common.GetConfigString("smtp_server") + ":" + common.GetConfigString("smtp_port")
	if err := smtp.SendMail(addr, auth, common.GetConfigString("smtp_username"), to, []byte(body)); err != nil {
		common.Error(err.Error())
		return err
	}
	return nil
}

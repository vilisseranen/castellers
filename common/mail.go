package common

import (
	"github.com/gadelkareem/go-helpers"
)

func SendMail(subject, body string, to []string) error {
	return h.SendMail(GetConfigString("smtp_server"), GetConfigString("mail_from"), subject, body, to)
}

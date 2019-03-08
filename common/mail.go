package common

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var mailService *gmail.Service

func init() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	mailService, err = gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		log.Fatalf("Unable to retrieve token from file: %v", err)
	}
	return config.Client(context.Background(), tok)
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func SendRegistrationEmail(to, memberName, language, adminName, adminExtra, activateLink, profileLink string) error {
	var msg gmail.Message
	// Prepare header
	header := "Subject: Inscription\r\n" +
		"To: " + to + "\r\n" +
		"From: Castellers de Montréal <" + GetConfigString("mail_from") + ">\r\n" +
		"Reply-To: " + GetConfigString("reply_to") + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
		"\r\n"
	// Parse body
	t, err := template.ParseFiles("templates/email_register_" + language + ".html")
	if err != nil {
		fmt.Println("Error parsing template: " + err.Error())
		return err
	}
	buf := new(bytes.Buffer)
	buf.WriteString(header)
	imageSource := GetConfigString("domain") + "/static/img/"
	emailInfo := emailRegisterInfo{memberName, adminName, adminExtra, activateLink, profileLink, imageSource}
	if err = t.Execute(buf, emailInfo); err != nil {
		fmt.Println("Error generating template: " + err.Error())
		return err
	}

	msg.Raw = base64.URLEncoding.EncodeToString(buf.Bytes())

	// Send the message
	_, err = mailService.Users.Messages.Send("me", &msg).Do()
	// Send mail
	if err != nil {
		fmt.Println("Error sending Email: " + err.Error())
		return err
	}
	return nil
}

type emailRegisterInfo struct {
	MemberName             string
	AdminName, AdminExtra  string
	LoginLink, ProfileLink string
	ImageSource            string
}

type email struct {
	from    string
	to      []string
	subject string
	body    string
}

func sendMail(to []string, body string) error {
	var auth smtp.Auth
	auth = smtp.PlainAuth("", GetConfigString("smtp_username"), GetConfigString("smtp_password"), GetConfigString("smtp_server"))
	addr := GetConfigString("smtp_server") + ":" + GetConfigString("smtp_port")
	if err := smtp.SendMail(addr, auth, GetConfigString("mail_from"), to, []byte(body)); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

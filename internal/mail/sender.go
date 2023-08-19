package mail

import (
	"fmt"
	simplemail "github.com/xhit/go-simple-mail"
	"os"
	"strings"
	"time"
)

const (
	serverHost           = "localhost"
	serverPort           = 1025
	serverConnectTimeout = 10 * time.Second
	serverKeepAlive      = false
)

type EmailSender interface {
	SendEmail(data Data) error
}

type HogSender struct {
	fromEmailAddress string
}

func NewHogSender(fromEmailAddress string) EmailSender {
	return &HogSender{
		fromEmailAddress: fromEmailAddress,
	}
}

type AttachFile struct {
	Name string
	Path string
}

type Data struct {
	To       []string
	Subject  string
	Content  string
	Files    []AttachFile
	Template string
}

// SendEmail sends an email
func (sender *HogSender) SendEmail(data Data) error {
	server := simplemail.NewSMTPClient()
	server.Host = serverHost
	server.Port = serverPort
	server.KeepAlive = serverKeepAlive
	server.ConnectTimeout = serverConnectTimeout

	client, err := server.Connect()
	if err != nil {
		return err
	}
	email := simplemail.NewMSG()
	email.SetFrom(sender.fromEmailAddress).
		SetSubject(data.Subject)

	// add To
	for _, t := range data.To {
		email.AddTo(t)
	}

	// attach files
	for _, f := range data.Files {
		email.AddAttachment(f.Path, f.Name)
	}

	// set body - if template is empty, use plain text, otherwise use provided email template
	if data.Template == "" {
		email.SetBody(simplemail.TextHTML, data.Content)
	} else {
		emailData, err := os.ReadFile(fmt.Sprintf("internal/mail/email_templates/%s", data.Template))
		if err != nil {
			return err
		}

		mailTemplate := string(emailData)
		finalMsg := strings.Replace(mailTemplate, "[%body%]", data.Content, 1)
		email.SetBody(simplemail.TextHTML, finalMsg)
	}

	err = email.Send(client)
	if err != nil {
		return err
	}

	return nil
}

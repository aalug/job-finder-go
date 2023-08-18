package mail

import (
	simplemail "github.com/xhit/go-simple-mail"
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
	To      []string
	Subject string
	Content string
	Files   []AttachFile
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

	email.SetBody(simplemail.TextHTML, data.Content)

	err = email.Send(client)
	if err != nil {
		return err
	}

	return nil
}

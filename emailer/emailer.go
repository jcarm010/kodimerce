package emailer

import (
	"github.com/jcarm010/kodimerce/log"
	"github.com/jcarm010/kodimerce/settings"
	"golang.org/x/net/context"
	"gopkg.in/gomail.v2"
	"gopkg.in/sendgrid/sendgrid-go.v2"
	"net/http"
	"strings"
	"time"
)

func SendGridEmail(ctx context.Context, from string, to string, subject string, body string, bcc string) error {
	generalSettings := settings.GetGlobalSettings(ctx)
	toEmails := strings.Split(to, ",")
	bccEmails := strings.Split(bcc, ",")
	sg := sendgrid.NewSendGridClientWithApiKey(generalSettings.SendGridKey)
	sg.Client = &http.Client{
		Timeout: time.Second * 10,
	}
	message := sendgrid.NewMail()
	for _, email := range toEmails {
		message.AddTo(strings.TrimSpace(email))
	}
	message.SetSubject(subject)
	message.SetHTML(body)
	message.SetFrom(from)
	for _, bcc := range bccEmails {
		bcc = strings.TrimSpace(bcc)
		if bcc == "" {
			continue
		}

		message.AddBcc(bcc)
	}

	return sg.Send(message)
}

func SendSMTPEmail(ctx context.Context, from string, to string, subject string, body string, bcc string) error {
	generalSettings := settings.GetGlobalSettings(ctx)
	toEmails := strings.Split(to, ",")
	bccEmails := strings.Split(bcc, ",")
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", toEmails...)
	if bcc != "" {
		m.SetHeader("Bcc", bccEmails...)
	}

	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	log.Infof(ctx, "toEmails: %s", toEmails)
	log.Infof(ctx, "len(toEmails): %s", len(toEmails))
	log.Infof(ctx, "generalSettings.SMTPUserName: %s", generalSettings.SMTPUserName)
	log.Infof(ctx, "generalSettings.SMTPPassword: %s", generalSettings.SMTPPassword)
	d := gomail.NewDialer(generalSettings.SMTPServer, generalSettings.SMTPPort, generalSettings.SMTPUserName, generalSettings.SMTPPassword)
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func SendEmail(ctx context.Context, from string, to string, subject string, body string, bcc string) error {
	generalSettings := settings.GetGlobalSettings(ctx)
	if generalSettings.SendGridKey != "" {
		return SendGridEmail(ctx, from, to, subject, body, bcc)
	} else {
		return SendSMTPEmail(ctx, from, to, subject, body, bcc)
	}
}

package emailer

import (
	"golang.org/x/net/context"
	"github.com/jcarm010/kodimerce/settings"
	"gopkg.in/sendgrid/sendgrid-go.v2"
	"google.golang.org/appengine/urlfetch"
	"strings"
	"google.golang.org/appengine/mail"
)

func SendEmail(ctx context.Context, from string, to string, subject string, body string, bcc string) error {
	generalSettings := settings.GetGlobalSettings(ctx)
	toEmails := strings.Split(to, ",")
	bccEmails := strings.Split(bcc, ",")
	if generalSettings.SendGridKey != "" {
		sg := sendgrid.NewSendGridClientWithApiKey(generalSettings.SendGridKey)
		sg.Client = urlfetch.Client(ctx)
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
	} else {
		msg := &mail.Message{
			Sender:  from,
			To:      toEmails,
			Subject: subject,
			Bcc:     bccEmails,
			HTMLBody: body,
		}

		return  mail.Send(ctx, msg)
	}
}

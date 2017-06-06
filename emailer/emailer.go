package emailer

import (
	"golang.org/x/net/context"
	"github.com/jcarm010/kodimerce/settings"
	"gopkg.in/sendgrid/sendgrid-go.v2"
	"google.golang.org/appengine/urlfetch"
	"strings"
)

func SendEmail(ctx context.Context, from string, to string, subject string, body string, bcc []string) error {
	toEmails := strings.Split(to, ",")
	sg := sendgrid.NewSendGridClientWithApiKey(settings.SENDGRID_KEY)
	sg.Client = urlfetch.Client(ctx)
	message := sendgrid.NewMail()
	for _, email := range toEmails {
		message.AddTo(strings.TrimSpace(email))
	}
	message.SetSubject(subject)
	message.SetHTML(body)
	message.SetFrom(from)
	for _, bcc := range bcc {
		message.AddBcc(bcc)
	}

	return sg.Send(message)
}

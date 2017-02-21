package emailer

import (
	"golang.org/x/net/context"
	"settings"
	"gopkg.in/sendgrid/sendgrid-go.v2"
	"google.golang.org/appengine/urlfetch"
)

func SendEmail(ctx context.Context, from string, to string, subject string, body string) error {
	sg := sendgrid.NewSendGridClientWithApiKey(settings.SENDGRID_KEY)
	sg.Client = urlfetch.Client(ctx)
	message := sendgrid.NewMail()
	message.AddTo(to)
	message.SetSubject(subject)
	message.SetHTML(body)
	message.SetFrom(from)
	return sg.Send(message)
}

package emailhelper

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/sethvargo/go-envconfig"
	"gopkg.in/gomail.v2"
)

type ActivationMailDriver interface {
	SendActivationEmail(email string, activationToken string, validUntil time.Time) error
}

func NewActivationMailHelper(ctx context.Context) (ActivationMailDriver, error) {
	var email activationMailDriverImpl

	if err := envconfig.Process(ctx, &email); err != nil {
		return nil, err
	}
	email.Dialer = gomail.NewDialer(email.EmailHost, email.EmailPort, email.EmailUser, email.EmailPass)

	return email, nil
}

type activationMailDriverImpl struct {
	Dialer    *gomail.Dialer
	Host      string `env:"HOST, required"`
	EmailHost string `env:"SMTP_HOST, required"`
	EmailPort int    `env:"SMTP_PORT, required"`
	EmailUser string `env:"SMTP_USER, required"`
	EmailPass string `env:"SMTP_PASS, required"`
	EmailFrom string `env:"SMTP_EMAIL, required"`
}

func (dialMail activationMailDriverImpl) SendActivationEmail(email string, activationToken string, validUntil time.Time) error {
	mailSetup := gomail.NewMessage()
	activationLink := fmt.Sprintf(
		"%s/auth/act?email=%s&token=%s",
		dialMail.Host,
		url.QueryEscape(email),
		url.QueryEscape(activationToken),
	)

	mailSetup.SetHeader("From", dialMail.EmailFrom)
	mailSetup.SetHeader("To", email)
	mailSetup.SetHeader("Subject", "Activate your account")
	mailSetup.SetBody("text/html",
		"<p>Thank you for registering. To complete, click the link below.</p>"+
			fmt.Sprintf("<p><a href=\"%s\">Click here</a></p>", activationLink)+
			"<p>--eLibrary--</p>",
	)

	return dialMail.Dialer.DialAndSend(mailSetup)
}

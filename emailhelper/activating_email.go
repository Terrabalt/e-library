package emailhelper

import (
	"fmt"
	"time"

	"gopkg.in/gomail.v2"
)

type ActivationMailDriver interface {
	SendActivationEmail(email string, activationToken string, validUntil time.Time) error
}

type ActivationMailDriverInst struct {
	Dialer    *gomail.Dialer
	Host      string `env:"HOST, required"`
	EmailHost string `env:"SMTP_HOST, required"`
	EmailPort int    `env:"SMTP_PORT, required"`
	EmailUser string `env:"SMTP_USER, required"`
	EmailPass string `env:"SMTP_PASS, required"`
	EmailFrom string `env:"SMTP_EMAIL, required"`
}

func (dialMail ActivationMailDriverInst) SendActivationEmail(email string, activationToken string, validUntil time.Time) error {
	mailSetup := gomail.NewMessage()

	mailSetup.SetHeader("From", dialMail.EmailFrom)
	mailSetup.SetHeader("To", email)
	mailSetup.SetHeader("Subject", "Activate your account")
	mailSetup.SetBody("text/html",
		"<p>Thank you for registering. To complete, click the link below.</p>"+
			fmt.Sprintf("<p><a href=\"%s/auth/act?token=%s\">Click here</a></p>", dialMail.Host, activationToken)+
			"<p>--eLibrary--</p>",
	)

	return dialMail.Dialer.DialAndSend(mailSetup)
}

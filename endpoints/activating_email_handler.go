package endpoints

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/gomail.v2"
)

func SendActivationEmail(w http.ResponseWriter, r *http.Request, email string, activationToken string, validUntil time.Time) error {
	mailSetup := gomail.NewMessage()

	mailSetup.SetHeader("From", os.Getenv("SMTP_EMAIL"))
	mailSetup.SetHeader("To", email)
	mailSetup.SetBody("text/html",
		"Thank you for registering. To complete, click the link below.\n"+
			fmt.Sprintf("<a href=\"http://%s/auth/act?token=%s\">Click here</a>\n", os.Getenv("HOST"), activationToken)+
			"--eLibrary--",
	)

	mailPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return err
	}

	dialMail := gomail.NewDialer(os.Getenv("SMTP_HOST"), mailPort, os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASSWORD"))
	return dialMail.DialAndSend(mailSetup)
}

package endpoints

import (
	"net/http"
	"time"
)

func SendActivationEmail(w http.ResponseWriter, r *http.Request, activationToken string, validUntil time.Time) {

}

package endpoints

import (
	"net/http"

	"github.com/go-chi/render"
)

type ErrorResponse struct {
	err            error  `json:"-"`       // low-level runtime error
	httpStatusCode int    `json:"-"`       // http response status code
	Message        string `json:"message"` // user-level status message
}

func (er *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, er.httpStatusCode)
	w.Header().Set("content-type", "application/json")
	return nil
}

func BadRequestError(err error) render.Renderer {
	return &ErrorResponse{
		err:            err,
		httpStatusCode: http.StatusBadRequest,
		Message:        "Bad Request - " + err.Error(),
	}
}

func UnauthorizedRequestError(err error) render.Renderer {
	return &ErrorResponse{
		err:            err,
		httpStatusCode: http.StatusUnauthorized,
		Message:        "Unauthorized Request - " + err.Error(),
	}
}

func ValidationFailedError(err error) render.Renderer {
	return &ErrorResponse{
		err:            err,
		httpStatusCode: http.StatusUnprocessableEntity,
		Message:        "Validation Failed - " + err.Error(),
	}
}

func InternalServerError(err error) render.Renderer {
	return &ErrorResponse{
		err:            err,
		httpStatusCode: http.StatusInternalServerError,
		Message:        "Internal Server Error - " + err.Error(),
	}
}

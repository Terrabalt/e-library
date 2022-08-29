package endpoints

import (
	"net/http"

	"github.com/go-chi/render"
)

type ErrorResponse struct {
	httpStatusCode int    `json:"-"`
	ErrorType      string `json:"error_type"`
	Message        string `json:"message"`
}

func (er *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, er.httpStatusCode)
	w.Header().Set("content-type", "application/json")
	return nil
}

func BadRequestError(err error) render.Renderer {
	return &ErrorResponse{
		httpStatusCode: http.StatusBadRequest,
		ErrorType:      "Bad Request",
		Message:        err.Error(),
	}
}

func RequestConflictError(err error) render.Renderer {
	return &ErrorResponse{
		httpStatusCode: http.StatusConflict,
		ErrorType:      "Conflict",
		Message:        err.Error(),
	}
}

func UnauthorizedRequestError(err error) render.Renderer {
	return &ErrorResponse{
		httpStatusCode: http.StatusUnauthorized,
		ErrorType:      "Unauthorized Request",
		Message:        err.Error(),
	}
}

func ValidationFailedError(err error) render.Renderer {
	return &ErrorResponse{
		httpStatusCode: http.StatusUnprocessableEntity,
		ErrorType:      "Validation Failed",
		Message:        err.Error(),
	}
}

func InternalServerError() render.Renderer {
	return &ErrorResponse{
		httpStatusCode: http.StatusInternalServerError,
		ErrorType:      "Internal Server Error",
		Message:        "something went wrong",
	}
}

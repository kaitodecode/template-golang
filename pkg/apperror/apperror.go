// pkg/apperror/apperror.go
package apperror

import "net/http"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Detail     any    `json:"detail,omitempty"`
	Stack      string `json:"stack,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code, message string, status int, detail any, stack string) *AppError {
	return &AppError{Code: code, Message: message, StatusCode: status, Detail: detail, Stack: stack}
}

var (
	ErrBadRequest = New("BAD_REQUEST", "Request tidak valid", http.StatusBadRequest, nil, "")
	ErrInternal   = New("INTERNAL_ERROR", "Kesalahan server", http.StatusInternalServerError, nil, "")
)

package base

import "strings"

type AppError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code int, messages ...string) *AppError {
	var msg string
	if len(messages) > 0 {
		msg = strings.Join(messages, "")
	}
	return &AppError{
		Code:    code,
		Message: msg,
	}
}

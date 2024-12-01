package utils

import (
	"fmt"
)

type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("Message: %s, Code: %d", e.Message, e.Code)
}

func NewError(msg string, code int) Error {
	return Error{
		Message: msg,
		Code:    code,
	}
}

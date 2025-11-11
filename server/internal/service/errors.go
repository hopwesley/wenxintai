package service

import "fmt"

type ErrorCode string

const (
        ErrorCodeBadRequest ErrorCode = "BAD_REQUEST"
        ErrorCodeNotFound   ErrorCode = "NOT_FOUND"
        ErrorCodeConflict   ErrorCode = "CONFLICT"
        ErrorCodeInternal   ErrorCode = "INTERNAL"
)

type Error struct {
        Code    ErrorCode
        Message string
        Err     error
}

func (e *Error) Error() string {
        if e.Message != "" {
                return e.Message
        }
        if e.Err != nil {
                return e.Err.Error()
        }
        return string(e.Code)
}

func (e *Error) Unwrap() error {
        return e.Err
}

func newError(code ErrorCode, message string, err error) *Error {
        if message == "" {
                message = string(code)
                if err != nil {
                        message = fmt.Sprintf("%s: %v", code, err)
                }
        }
        return &Error{Code: code, Message: message, Err: err}
}

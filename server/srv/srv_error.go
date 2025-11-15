package srv

import (
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrorCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrorCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrorCodeConflict       ErrorCode = "CONFLICT"
	ErrorCodeInternal       ErrorCode = "INTERNAL"
	ErrorCodeInviteReserved ErrorCode = "RESERVED"
	ErrorCodeInviteDisabled ErrorCode = "DISABLED"
	ErrorCodeInviteRedeemed ErrorCode = "REDEEMED"
)

type ApiErr struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Err     error     `json:"err,omitempty"`
	status  int
}

var (
	ApiMethodInvalid = NewError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
)

func ApiInvalidReq(msg string, err error) *ApiErr {
	return NewError(http.StatusBadRequest, ErrorCodeBadRequest, msg, err)
}

func (e *ApiErr) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

func (e *ApiErr) Unwrap() error {
	return e.Err
}

func NewError(status int, code ErrorCode, message string, err error) *ApiErr {
	if message == "" {
		message = string(code)
		if err != nil {
			message = fmt.Sprintf("%s: %v", code, err)
		}
	}
	return &ApiErr{Code: code, Message: message, Err: err, status: status}
}

package srv

import (
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrorCodeBadRequest ErrorCode = "BAD_REQUEST"
	ErrorCodeNotFound   ErrorCode = "NOT_FOUND"
	ErrorCodeForbidden  ErrorCode = "FORBIDEEN"
	ErrorCodeInternal   ErrorCode = "INTERNAL"
	ErrorCodeSequence   ErrorCode = "BAD_SEQ"
)

type ApiErr struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Err     error     `json:"err,omitempty"`
	status  int
}

func ApiInvalidReq(msg string, err error) *ApiErr {
	return NewApiError(http.StatusBadRequest, ErrorCodeBadRequest, msg, err)
}

func ApiInternalErr(msg string, err error) *ApiErr {
	return NewApiError(http.StatusInternalServerError, ErrorCodeInternal, msg, err)
}

func ApiInvalidTestSequence(err error) *ApiErr {
	return NewApiError(http.StatusInternalServerError, ErrorCodeSequence, "请按照测试顺序进行测试", err)
}

func ApiInvalidNoTestRecord(err error) *ApiErr {
	return NewApiError(http.StatusInternalServerError, ErrorCodeNotFound, "未找到问卷数据", err)
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

func NewApiError(status int, code ErrorCode, message string, err error) *ApiErr {
	if message == "" {
		message = string(code)
		if err != nil {
			message = fmt.Sprintf("%s: %v", code, err)
		}
	}
	return &ApiErr{Code: code, Message: message, Err: err, status: status}
}

package comm

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
	ErrType     = errors.New("invalid type convert")
)

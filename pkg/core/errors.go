package core

import (
	"errors"
	"fmt"
)

// ErrorCode is a stable, machine-readable code for categorizing errors.
type ErrorCode string

const (
	// Common
	ErrCodeInvalidArgument ErrorCode = "INVALID_ARGUMENT"
	ErrCodeInternal        ErrorCode = "INTERNAL"

	// Driver (fetch)
	ErrCodeDriverBuildRequest   ErrorCode = "DRIVER_BUILD_REQUEST"
	ErrCodeDriverRequestFailed  ErrorCode = "DRIVER_REQUEST_FAILED"
	ErrCodeDriverUnexpectedHTTP ErrorCode = "DRIVER_UNEXPECTED_HTTP_STATUS"
	ErrCodeDriverReadBodyFailed ErrorCode = "DRIVER_READ_BODY_FAILED"
)

// CodedError wraps an underlying error with a stable error code.
type CodedError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *CodedError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Message == "" {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *CodedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewCodedError(code ErrorCode, message string, err error) error {
	return &CodedError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func ErrorCodeOf(err error) ErrorCode {
	if err == nil {
		return ""
	}
	var coded *CodedError
	if errors.As(err, &coded) {
		return coded.Code
	}
	return ""
}

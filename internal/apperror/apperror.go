package apperror

import (
	"fmt"
)

// AppError defines a standard application error.
type AppError struct {
	Type    string
	Message string
	Status  int
	Err     error
}

// NewAppError creates a new AppError.
func NewAppError(errType, message string, status int, err error) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Status:  status,
		Err:     err,
	}
}

// WithError will clone the existing error and set the Err field for extra details.
// This is used for already-initalized app errors where the error is not already provided.
func (e *AppError) WithError(err error) *AppError {
	c := e
	c.Err = err
	return c
}

// Error will return the error.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}

// Unwrap will unwrap the error.
func (e *AppError) Unwrap() error {
	return e.Err
}

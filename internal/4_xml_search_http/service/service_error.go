package service

import (
	"errors"
)

type TypeServiceError int

const (
	BadRequest TypeServiceError = iota
	InternalError
)

var ErrServiceError = errors.New("service error")

type ServiceError struct {
	TypeError TypeServiceError
	err       error
}

func (e *ServiceError) Error() string {
	return e.err.Error()
}

func (e *ServiceError) Unwrap() error {
	return ErrServiceError
}

func NewServiceError(typeError TypeServiceError, err error) *ServiceError {
	return &ServiceError{
		TypeError: typeError,
		err:       err,
	}
}

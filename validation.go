package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
)

const (
	ErrInvalidStatusCode failure.StringCode = "status-code"
	ErrInvalidPath       failure.StringCode = "path"
	ErrNotFound          failure.StringCode = "not-found"
	ErrCSRFToken         failure.StringCode = "csrf-token"
	ErrInvalidPostOrder  failure.StringCode = "post-order"
	ErrInvalidAsset      failure.StringCode = "asset"
)

type ValidationError struct {
	Errors []error
}

func (v ValidationError) Error() string {
	messages := []string{}
	for _, err := range v.Errors {
		if err != nil {
			messages = append(messages, fmt.Sprintf("%v", err))
		}
	}
	return strings.Join(messages, "\n")
}

func (v ValidationError) IsEmpty() bool {
	for _, err := range v.Errors {
		if err != nil {
			if ve, ok := err.(ValidationError); ok {
				if !ve.IsEmpty() {
					return false
				}
			} else {
				return false
			}
		}
	}
	return true
}

func (v ValidationError) Add(step *isucandar.BenchmarkStep) {
	for _, err := range v.Errors {
		if err != nil {
			if ve, ok := err.(ValidationError); ok {
				ve.Add(step)
			} else {
				step.AddError(err)
			}
		}
	}
}

type ResponseValidator func(*http.Response) error

func ValidateResponse(res *http.Response, validators ...ResponseValidator) ValidationError {
	errs := []error{}
	for _, validator := range validators {
		if err := validator(res); err != nil {
			errs = append(errs, err)
		}
	}
	return ValidationError{Errors: errs}
}

func WithStatusCode(statusCode int) ResponseValidator {
	return func(r *http.Response) error {
		if r.StatusCode != statusCode {
			return failure.NewError(
				ErrInvalidStatusCode,
				fmt.Errorf("%s %s : expected(%d) != actual(%d)", r.Request.Method, r.Request.URL.Path, statusCode, r.StatusCode),
			)
		}
		return nil
	}
}

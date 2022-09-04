package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
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

func WithIncludeBody(val string) ResponseValidator {
	return func(r *http.Response) error {
		defer r.Body.Close()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return failure.NewError(ErrInvalidResposne, fmt.Errorf("%s %s : %s", r.Request.Method, r.Request.URL.Path, err.Error()))
		}

		if bytes.IndexAny(body, val) == -1 {
			return failure.NewError(ErrNotFound, fmt.Errorf("%s %s : %s is not found in body", r.Request.Method, r.Request.URL.Path, val))
		}

		return nil
	}
}

func WithCSRFToken(user *User) ResponseValidator {
	return func(r *http.Response) error {
		defer r.Body.Close()

		user.SetCSRFToken("")

		doc, err := goquery.NewDocumentFromReader(r.Body)
		if err != nil {
			return failure.NewError(ErrInvalidResposne, fmt.Errorf("%s %s : %s", r.Request.Method, r.Request.URL.Path, err.Error()))
		}

		node := doc.Find(`input[name="csrf_token"]`).Get(0)
		if node == nil {
			return failure.NewError(ErrCSRFToken, fmt.Errorf("%s %s : CSRF token is not found", r.Request.Method, r.Request.URL.Path))
		}

		for _, attr := range node.Attr {
			if attr.Key == "value" {
				user.SetCSRFToken(attr.Val)
			}
		}

		if user.GetCSRFToken() == "" {
			return failure.NewError(ErrCSRFToken, fmt.Errorf("%s %s : CSRF token is not found", r.Request.Method, r.Request.URL.Path))
		}

		return nil
	}
}

func WithLocation(val string) ResponseValidator {
	return func(r *http.Response) error {
		target := r.Request.URL.ResolveReference(&url.URL{Path: val})
		if r.Header.Get("Location") != target.String() {
			return failure.NewError(
				ErrInvalidPath,
				fmt.Errorf("%s %s : %s, expected(%s) != actual(%s)", r.Request.Method, r.Request.URL.Path, "Location", val, r.Header.Get("Location")),
			)
		}
		return nil
	}
}

var (
	assetsMD5 = map[string]string{
		"favicon.ico":       "ad4b0f606e0f8465bc4c4c170b37e1a3",
		"js/timeago.min.js": "f2d4c53400d0a46de704f5a97d6d04fb",
		"js/main.js":        "9c309fed7e360c57a705978dab2c68ad",
		"css/style.css":     "e4c3606a18d11863189405eb5c6ca551",
	}
)

func WithAssets(ctx context.Context, ag *agent.Agent) ResponseValidator {
	return func(r *http.Response) error {
		resources, err := ag.ProcessHTML(ctx, r, r.Body)
		if err != nil {
			return failure.NewError(
				ErrInvalidAsset,
				fmt.Errorf("%s %s : %v", r.Request.Method, r.Request.URL.Path, err),
			)
		}

		errs := []error{}

		for uri, res := range resources {
			path := strings.TrimPrefix(uri, ag.BaseURL.String())
			if res.Error != nil {
				errs = append(errs, failure.NewError(ErrInvalidAsset, fmt.Errorf("%s / %s : %v", "GET", path, res.Error)))
				continue
			}

			defer res.Response.Body.Close()

			if res.Response.StatusCode == 304 {
				continue
			}

			expectedMD5, ok := assetsMD5[path]
			if !ok {
				continue
			}

			hash := md5.New()
			io.Copy(hash, res.Response.Body)
			actualMD5 := hex.EncodeToString(hash.Sum(nil))

			if expectedMD5 != actualMD5 {
				errs = append(errs, failure.NewError(ErrInvalidAsset, fmt.Errorf("%s / %s : expected(MD5 %s) != actual(MD5 %s)", "GET", path, expectedMD5, actualMD5)))
			}
		}
		return ValidationError{Errors: errs}
	}
}

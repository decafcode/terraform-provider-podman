package client

import (
	"fmt"
	"net/url"
)

type HttpError struct {
	URL *url.URL
}

type ContentTypeError struct {
	HttpError
	ContentType string
}

type StatusCodeError struct {
	HttpError
	StatusCode int
	Message    string
}

func (e ContentTypeError) Error() string {
	return fmt.Sprintf("%s: expected application/json response, got %s", e.URL, e.ContentType)
}

func (e StatusCodeError) Error() string {
	return fmt.Sprintf("%s: server returned status code %d: %s", e.URL, e.StatusCode, e.Message)
}

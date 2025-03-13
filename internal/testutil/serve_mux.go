package testutil

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type serveMux struct {
	http.ServeMux
	BaseURL *url.URL
	Timeout time.Duration
}

type handlerFunc func(context.Context, http.ResponseWriter, *http.Request) error

type handler interface {
	Handler(context.Context, http.ResponseWriter, *http.Request) error
}

type statusError struct {
	Code    int
	Message string
}

func (e statusError) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Message)
}

func (m *serveMux) HandleFunc(method string, pathPattern string, fn handlerFunc) {
	pattern := fmt.Sprintf("%s %s%s", method, m.BaseURL.Path, pathPattern)

	m.ServeMux.HandleFunc(pattern, func(resp http.ResponseWriter, req *http.Request) {
		var ctx context.Context
		var cancel context.CancelFunc

		if m.Timeout != 0 {
			ctx, cancel = context.WithTimeout(context.Background(), m.Timeout)
		} else {
			ctx, cancel = context.WithCancel(context.Background())
		}

		defer cancel()
		err := fn(ctx, resp, req.WithContext(ctx))

		if err != nil {
			statusError, ok := err.(statusError)

			if ok {
				http.Error(resp, statusError.Message, statusError.Code)
			} else {
				fmt.Printf("Unhandled error: %v", err)
				http.Error(resp, "internal server error", http.StatusInternalServerError)
			}
		}
	})
}

func (m *serveMux) Handle(method string, pathPattern string, h handler) {
	m.HandleFunc(method, pathPattern, h.Handler)
}

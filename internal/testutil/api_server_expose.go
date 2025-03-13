package testutil

import (
	"net/http"
	"net/url"
	"time"
)

func (s *ApiServer) Expose(baseURL *url.URL, timeout time.Duration) http.Handler {
	mux := &serveMux{BaseURL: baseURL, Timeout: timeout}

	mux.HandleFunc("GET", "v5.0.0/libpod/_ping", s.handlePing)
	mux.HandleFunc("POST", "v5.0.0/libpod/networks/create", s.handleNetworkCreate)
	mux.HandleFunc("DELETE", "v5.0.0/libpod/networks/{nameOrId}", s.handleNetworkDelete)
	mux.HandleFunc("GET", "v5.0.0/libpod/networks/{nameOrId}/json", s.handleNetworkGet)

	return mux
}

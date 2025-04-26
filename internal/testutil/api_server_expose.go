package testutil

import (
	"net/http"
	"net/url"
	"time"
)

func (s *ApiServer) Expose(baseURL *url.URL, timeout time.Duration) http.Handler {
	mux := &serveMux{BaseURL: baseURL, Timeout: timeout}

	mux.HandleFunc("GET", "v5.0.0/libpod/_ping", s.handlePing)
	mux.HandleFunc("POST", "v5.0.0/libpod/containers/create", s.handleContainerCreate)
	mux.HandleFunc("DELETE", "v5.0.0/libpod/containers/{nameOrId}", s.handleContainerDelete)
	mux.HandleFunc("GET", "v5.0.0/libpod/containers/{nameOrId}/json", s.handleContainerGet)
	mux.HandleFunc("POST", "v5.0.0/libpod/containers/{nameOrId}/rename", s.handleContainerRename)
	mux.HandleFunc("POST", "v5.0.0/libpod/containers/{nameOrId}/start", s.handleContainerStart)
	mux.HandleFunc("POST", "v5.0.0/libpod/containers/{nameOrId}/stop", s.handleContainerStop)
	mux.HandleFunc("POST", "v5.0.0/libpod/images/pull", s.handleImagePull)
	mux.HandleFunc("DELETE", "v5.0.0/libpod/images/{nameOrId}", s.handleImageDelete)
	mux.HandleFunc("GET", "v5.0.0/libpod/images/{nameOrId}/json", s.handleImageGet)
	mux.HandleFunc("POST", "v5.0.0/libpod/networks/create", s.handleNetworkCreate)
	mux.HandleFunc("DELETE", "v5.0.0/libpod/networks/{nameOrId}", s.handleNetworkDelete)
	mux.HandleFunc("GET", "v5.0.0/libpod/networks/{nameOrId}/json", s.handleNetworkGet)
	mux.HandleFunc("POST", "v5.0.0/libpod/secrets/create", s.handleSecretCreate)
	mux.HandleFunc("DELETE", "v5.0.0/libpod/secrets/{nameOrId}", s.handleSecretDelete)
	mux.HandleFunc("GET", "v5.0.0/libpod/secrets/{nameOrId}/json", s.handleSecretGet)

	return mux
}

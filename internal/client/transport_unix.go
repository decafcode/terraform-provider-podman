package client

import (
	"context"
	"net"
	"net/http"
)

type unixTransport struct {
	net.Dialer
	unixSocket string
}

func createUnixTransport(path string) *http.Client {
	transport := &unixTransport{
		unixSocket: path,
	}

	http := &http.Client{
		Transport: &http.Transport{
			DialContext: transport.DialContext,
		},
	}

	return http
}

func (t *unixTransport) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return t.Dialer.DialContext(ctx, "unix", t.unixSocket)
}

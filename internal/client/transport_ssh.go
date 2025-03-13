package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/crypto/ssh"
)

type sshTransport struct {
	ssh        *ssh.Client
	unixSocket string
}

func dialSshTransport(ctx context.Context, url *url.URL, config *ssh.ClientConfig) (*sshTransport, *http.Client, error) {
	port := "22"

	if url.Port() != "" {
		port = url.Port()
	}

	dialer := net.Dialer{}
	addr := net.JoinHostPort(url.Hostname(), port)
	tcpConn, err := dialer.DialContext(ctx, "tcp", addr)

	if err != nil {
		return nil, nil, err
	}

	configCopy := *config
	username := url.User.Username()

	if username == "" {
		return nil, nil, fmt.Errorf("username is required when using SSH transport")
	}

	configCopy.User = username

	password, hasPassword := url.User.Password()

	if hasPassword {
		configCopy.Auth = append(configCopy.Auth, ssh.Password(password))
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(tcpConn, addr, &configCopy)

	if err != nil {
		return nil, nil, err
	}

	transport := &sshTransport{
		ssh:        ssh.NewClient(sshConn, chans, reqs),
		unixSocket: url.Path,
	}

	http := &http.Client{
		Transport: &http.Transport{
			DialContext: transport.DialContext,
		},
	}

	return transport, http, nil
}

func (t *sshTransport) Close() error {
	return t.ssh.Close()
}

func (t *sshTransport) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return t.ssh.DialContext(ctx, "unix", t.unixSocket)
}

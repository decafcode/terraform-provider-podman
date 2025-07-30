package provider

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/decafcode/terraform-provider-podman/internal/testutil"
	"golang.org/x/crypto/ssh"
)

type framework struct {
	port       net.Listener
	httpServer *http.Server
	clientUrl  string
}

func spawnFramework(_ context.Context, apiServer *testutil.ApiServer) (*framework, error) {
	clientUrl, err := url.Parse("tcp://localhost:55550/subpath/")

	if err != nil {
		panic(err)
	}

	port, err := net.Listen("tcp", clientUrl.Host)

	if err != nil {
		return nil, err
	}

	httpServer := &http.Server{Handler: apiServer.Expose(clientUrl, 1*time.Second)}
	go httpServer.Serve(port) // nolint:errcheck

	return &framework{
		port:       port,
		httpServer: httpServer,
		clientUrl:  clientUrl.String(),
	}, nil
}

func (f *framework) Stop(ctx context.Context) {
	f.httpServer.Shutdown(ctx) // nolint:errcheck
	f.port.Close()
}

func (f *framework) Url() string {
	return f.clientUrl
}

type sshFramework struct {
	port       net.Listener
	httpServer *http.Server
	sshServer  *testutil.SshServer
}

const sshClientUrl = "ssh://user:trustno1@localhost:55551/tmp/imaginary-socket"

func spawnSshFramework(_ context.Context, hostPrivateKey ssh.Signer, apiServer *testutil.ApiServer) (*sshFramework, error) {
	clientUrl, err := url.Parse(sshClientUrl)

	if err != nil {
		panic(err)
	}

	serverUrl, err := url.Parse("http://_d/")

	if err != nil {
		panic(err)
	}

	f := &sshFramework{}
	f.port, err = net.Listen("tcp", clientUrl.Host)

	if err != nil {
		return nil, err
	}

	password, _ := clientUrl.User.Password()

	f.sshServer = &testutil.SshServer{
		SocketPath:     clientUrl.Path,
		HostPrivateKey: hostPrivateKey,
		Password:       &password,
	}

	f.httpServer = &http.Server{
		Handler: apiServer.Expose(serverUrl, 1*time.Second),
	}

	go f.sshServer.Serve(f.port)       // nolint:errcheck
	go f.httpServer.Serve(f.sshServer) // nolint:errcheck

	return f, nil
}

func (f *sshFramework) Stop(ctx context.Context) {
	if f.httpServer != nil {
		f.httpServer.Shutdown(ctx) // nolint:errcheck
	}

	if f.port != nil {
		f.port.Close()
	}
}

func (f *sshFramework) Url() string {
	return sshClientUrl
}

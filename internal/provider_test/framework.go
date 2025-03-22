package provider

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/decafcode/terraform-provider-podman/internal/testutil"
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

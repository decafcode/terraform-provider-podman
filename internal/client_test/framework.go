package client

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/decafcode/terraform-provider-podman/internal/client"
	"github.com/decafcode/terraform-provider-podman/internal/testutil"
)

type framework struct {
	*client.Client
	port       net.Listener
	httpServer *http.Server
}

func spawnFramework(ctx context.Context, apiServer *testutil.ApiServer) (*framework, error) {
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

	client, err := client.Connect(ctx, clientUrl, nil)

	if err != nil {
		err2 := httpServer.Shutdown(ctx)

		if err2 != nil {
			panic(err2)
		}

		err2 = port.Close()

		if err2 != nil {
			panic(err2)
		}
	}

	return &framework{
		port:       port,
		httpServer: httpServer,
		Client:     client,
	}, nil
}

func (f *framework) Stop(ctx context.Context) {
	f.Client.Close()
	err := f.httpServer.Shutdown(ctx)
	f.port.Close()

	if err != nil {
		panic(err)
	}
}

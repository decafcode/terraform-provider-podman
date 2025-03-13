package client

import (
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/decafcode/terraform-provider-podman/internal/client"
	"github.com/decafcode/terraform-provider-podman/internal/testutil"
	"gotest.tools/v3/assert"
)

func TestHttpTransport(t *testing.T) {
	clientUrl, err := url.Parse("tcp://localhost:55550/subpath/")
	assert.NilError(t, err)

	port, err := net.Listen("tcp", clientUrl.Host)
	assert.NilError(t, err)
	defer port.Close()

	apiServer := &testutil.ApiServer{}
	httpServer := http.Server{Handler: apiServer.Expose(clientUrl, 1*time.Second)}

	go httpServer.Serve(port) // nolint:errcheck
	defer httpServer.Shutdown(t.Context())

	c, err := client.Connect(t.Context(), clientUrl, nil)
	assert.NilError(t, err)
	defer c.Close()

	err = c.Ping(t.Context())
	assert.NilError(t, err)
}

func TestUnixTransport(t *testing.T) {
	clientUrl, err := url.Parse("unix:///tmp/test-socket")
	assert.NilError(t, err)

	port, err := net.Listen("unix", clientUrl.Path)
	assert.NilError(t, err)
	defer port.Close()

	serverUrl, err := url.Parse("http://_d/")
	assert.NilError(t, err)

	apiServer := &testutil.ApiServer{}
	httpServer := http.Server{Handler: apiServer.Expose(serverUrl, 1*time.Second)}

	go httpServer.Serve(port) // nolint:errcheck
	defer httpServer.Shutdown(t.Context())

	c, err := client.Connect(t.Context(), clientUrl, nil)
	assert.NilError(t, err)

	err = c.Ping(t.Context())
	assert.NilError(t, err)
}

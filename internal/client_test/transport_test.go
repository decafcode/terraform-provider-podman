package client

import (
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/decafcode/terraform-provider-podman/internal/client"
	"github.com/decafcode/terraform-provider-podman/internal/testutil"
	"golang.org/x/crypto/ssh"
	"gotest.tools/v3/assert"
)

var hostPrivateKey, hostPrivateKeyErr = ssh.ParsePrivateKey([]byte(`
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBGoSgRlJh1AODGYCd6YuL2TH4AdSl24d91hMVIevPCYQAAAJCuxaQZrsWk
GQAAAAtzc2gtZWQyNTUxOQAAACBGoSgRlJh1AODGYCd6YuL2TH4AdSl24d91hMVIevPCYQ
AAAEA71OqHFuueBBi+9zt8UmiDbVxONw5FzxxJtaLzmwip0kahKBGUmHUA4MZgJ3pi4vZM
fgB1KXbh33WExUh688JhAAAAC3RhdUB0b29sYm94AQI=
-----END OPENSSH PRIVATE KEY-----
`))

var hostPublicKey, _, _, _, hostPublicKeyErr = ssh.ParseAuthorizedKey([]byte(`
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEahKBGUmHUA4MZgJ3pi4vZMfgB1KXbh33WExUh688Jh
`))

var privateKey, privateKeyErr = ssh.ParsePrivateKey([]byte(`
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBLhQ2W5ecrOySZUWDwkowIj8Ka5DTEu1M4NsWCNBlkdwAAAJAAfQppAH0K
aQAAAAtzc2gtZWQyNTUxOQAAACBLhQ2W5ecrOySZUWDwkowIj8Ka5DTEu1M4NsWCNBlkdw
AAAECTVsE1GD1kfD6OOUVuWCPsYs0TbGjnXihLWm6sQ4IZq0uFDZbl5ys7JJlRYPCSjAiP
wprkNMS7Uzg2xYI0GWR3AAAAC3RhdUB0b29sYm94AQI=
-----END OPENSSH PRIVATE KEY-----
`))

var publicKey, _, _, _, publicKeyErr = ssh.ParseAuthorizedKey([]byte(`
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEuFDZbl5ys7JJlRYPCSjAiPwprkNMS7Uzg2xYI0GWR3
`))

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

func TestSshTransportPassword(t *testing.T) {
	assert.NilError(t, hostPrivateKeyErr)
	assert.NilError(t, hostPublicKeyErr)

	clientUrl, err := url.Parse("ssh://user:trustno1@localhost:55551/tmp/imaginary-socket")
	assert.NilError(t, err)

	port, err := net.Listen("tcp", clientUrl.Host)
	assert.NilError(t, err)
	defer port.Close()

	password, _ := clientUrl.User.Password()
	sshServer := &testutil.SshServer{
		SocketPath:     clientUrl.Path,
		HostPrivateKey: hostPrivateKey,
		Password:       &password,
	}

	go sshServer.Serve(port) // nolint:errcheck
	defer sshServer.Close()

	serverUrl, err := url.Parse("http://_d/")
	assert.NilError(t, err)

	apiServer := &testutil.ApiServer{}
	httpServer := http.Server{Handler: apiServer.Expose(serverUrl, 1*time.Second)}

	go httpServer.Serve(sshServer)         // nolint:errcheck
	defer httpServer.Shutdown(t.Context()) // nolint:errcheck

	config := &client.Config{
		Ssh: ssh.ClientConfig{
			HostKeyCallback: ssh.FixedHostKey(hostPublicKey),
		},
	}

	c, err := client.Connect(t.Context(), clientUrl, config)
	assert.NilError(t, err)

	err = c.Ping(t.Context())
	assert.NilError(t, err)
}

func TestSshTransportPublicKey(t *testing.T) {
	assert.NilError(t, hostPrivateKeyErr)
	assert.NilError(t, hostPublicKeyErr)
	assert.NilError(t, privateKeyErr)
	assert.NilError(t, publicKeyErr)

	clientUrl, err := url.Parse("ssh://user@localhost:55551/tmp/imaginary-socket")
	assert.NilError(t, err)

	port, err := net.Listen("tcp", clientUrl.Host)
	assert.NilError(t, err)
	defer port.Close()

	sshServer := &testutil.SshServer{
		SocketPath:     clientUrl.Path,
		HostPrivateKey: hostPrivateKey,
		PublicKey:      publicKey,
	}

	go sshServer.Serve(port) // nolint:errcheck
	defer sshServer.Close()

	serverUrl, err := url.Parse("http://_d/")
	assert.NilError(t, err)

	apiServer := &testutil.ApiServer{}
	httpServer := http.Server{Handler: apiServer.Expose(serverUrl, 1*time.Second)}

	go httpServer.Serve(sshServer)         // nolint:errcheck
	defer httpServer.Shutdown(t.Context()) // nolint:errcheck

	config := &client.Config{
		Ssh: ssh.ClientConfig{
			HostKeyCallback: ssh.FixedHostKey(hostPublicKey),
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(privateKey),
			},
		},
	}

	c, err := client.Connect(t.Context(), clientUrl, config)
	assert.NilError(t, err)

	err = c.Ping(t.Context())
	assert.NilError(t, err)
}

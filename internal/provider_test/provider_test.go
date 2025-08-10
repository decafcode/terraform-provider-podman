package provider

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/decafcode/terraform-provider-podman/internal/provider"
	"github.com/decafcode/terraform-provider-podman/internal/testutil"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"golang.org/x/crypto/ssh"
	"gotest.tools/v3/assert"
)

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"podman": providerserver.NewProtocol6WithError(
		provider.New(
			"test",
			&provider.PodmanProviderEnv{},
		)(),
	),
}

func TestAccSshCommunication(t *testing.T) {
	caPublicKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOzMB/yLNO/Kd8qCrKpBp2Gd4MYb3ZdqK17wxbkDqJpO"
	hostPublicKey := "ssh-ed25519-cert-v01@openssh.com AAAAIHNzaC1lZDI1NTE5LWNlcnQtdjAxQG9wZW5zc2guY29tAAAAIE3iQn89N+NB66PoomeEammZAQLGf8xnSiCYYfKDziQSAAAAIEahKBGUmHUA4MZgJ3pi4vZMfgB1KXbh33WExUh688JhAAAAAAAAAAAAAAACAAAACWxvY2FsaG9zdAAAAAAAAAAAAAAAAP//////////AAAAAAAAAAAAAAAAAAAAMwAAAAtzc2gtZWQyNTUxOQAAACDszAf8izTvynfKgqyqQadhneDGG92Xaite8MW5A6iaTgAAAFMAAAALc3NoLWVkMjU1MTkAAABArOaKykP6KRrbtOarRqhRoJjS1RGSOSBg9of30y9E8y3RKvpE3WGp5JZnuael3vUqAsAVQ+RGFgCxhdj8jRgPBQ=="
	othrPublicKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEuFDZbl5ys7JJlRYPCSjAiPwprkNMS7Uzg2xYI0GWR3"
	hostPrivateKey, err := ssh.ParsePrivateKey([]byte(`
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBGoSgRlJh1AODGYCd6YuL2TH4AdSl24d91hMVIevPCYQAAAJCuxaQZrsWk
GQAAAAtzc2gtZWQyNTUxOQAAACBGoSgRlJh1AODGYCd6YuL2TH4AdSl24d91hMVIevPCYQ
AAAEA71OqHFuueBBi+9zt8UmiDbVxONw5FzxxJtaLzmwip0kahKBGUmHUA4MZgJ3pi4vZM
fgB1KXbh33WExUh688JhAAAAC3RhdUB0b29sYm94AQI=
-----END OPENSSH PRIVATE KEY-----
	`))

	assert.NilError(t, err)

	pubKeyBytes, err := base64.StdEncoding.DecodeString(strings.Split(hostPublicKey, " ")[1])
	assert.NilError(t, err)

	hostPublicKeyObj, err := ssh.ParsePublicKey(pubKeyBytes)
	assert.NilError(t, err)

	cert, ok := hostPublicKeyObj.(*ssh.Certificate)
	assert.Equal(t, ok, true)

	hostSigner, err := ssh.NewCertSigner(cert, hostPrivateKey)
	assert.NilError(t, err)

	apiServer := testutil.ApiServer{}
	f, err := spawnSshFramework(t.Context(), hostSigner, &apiServer)
	assert.NilError(t, err)

	defer f.Stop(t.Context())

	// Create an arbitrary resource to test communication with the server. The
	// exact type of resource does not matter. We'll use podman_network
	// because it's easy to work with.

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "podman_network" "test1" {
						container_host = "%s"
						name           = "test1"
					}
				`, f.Url()),
				// SSH URLs need to contain some sort of policy for dealing
				// with host keys in their URL fragments. Check that a
				// relevant error is being returned if this is missing.
				ExpectError: regexp.MustCompile("ssh container_host URL"),
			},
			{
				// Verify blind trust
				Config: fmt.Sprintf(`
					resource "podman_network" "test2" {
						container_host = "%s#trust_unknown_host=1"
						name           = "test2"
					}
				`, f.Url()),
			},
			{
				// Verify host key verification by checking for the specific
				// certificate being used as the host's public key
				Config: fmt.Sprintf(`
					resource "podman_network" "test3" {
						container_host = "%s#pubkey=%s"
						name           = "test3"
					}
				`, f.Url(), url.QueryEscape(hostPublicKey)),
			},
			{
				// Ensure that the host key is actually verified
				Config: fmt.Sprintf(`
					resource "podman_network" "test4" {
						container_host = "%s#pubkey=%s"
						name           = "test4"
					}
				`, f.Url(), url.QueryEscape(othrPublicKey)),
				ExpectError: regexp.MustCompile("public key mismatch"),
			},
			{
				// Trust anything signed by the CA that signed our test host's
				// certificate, not just that specific certificate.
				Config: fmt.Sprintf(`
					resource "podman_network" "test5" {
						container_host = "%s#ca=%s"
						name           = "test5"
					}
				`, f.Url(), url.QueryEscape(caPublicKey)),
			},
			{
				// Check that the CA signature is actually being verified
				Config: fmt.Sprintf(`
					resource "podman_network" "test6" {
						container_host = "%s#ca=%s"
						name           = "test6"
					}
				`, f.Url(), url.QueryEscape(othrPublicKey)),
				ExpectError: regexp.MustCompile("no authorities for hostname"),
			},
		},
	})
}

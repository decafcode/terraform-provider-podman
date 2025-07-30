package provider

import (
	"fmt"
	"net/url"
	"regexp"
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
	hostPublicKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEahKBGUmHUA4MZgJ3pi4vZMfgB1KXbh33WExUh688Jh"
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

	apiServer := testutil.ApiServer{}
	f, err := spawnSshFramework(t.Context(), hostPrivateKey, &apiServer)
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
				// Verify host key verification
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
		},
	})
}

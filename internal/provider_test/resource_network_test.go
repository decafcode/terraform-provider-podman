package provider

import (
	"fmt"
	"testing"

	"github.com/decafcode/terraform-provider-podman/internal/api"
	"github.com/decafcode/terraform-provider-podman/internal/provider"
	"github.com/decafcode/terraform-provider-podman/internal/testutil"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"gotest.tools/v3/assert"
)

func TestAccNetworkResource(t *testing.T) {
	n := &api.NetworkJson{
		Id:   "xyz",
		Name: "importtest",
	}

	apiServer := testutil.ApiServer{
		Networks: map[string]*api.NetworkJson{
			n.Id: n,
		},
	}

	framework, err := spawnFramework(t.Context(), &apiServer)
	assert.NilError(t, err)

	defer framework.Stop(t.Context())

	env := provider.PodmanProviderEnv{}

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"podman": providerserver.NewProtocol6WithError(provider.New("test", &env)()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "podman_network" "test" {
						container_host = "%s"
						name           = "mynetwork"
					}
				`, framework.Url()),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"podman_network.test",
						tfjsonpath.New("id"),
						isString{},
					),
					statecheck.ExpectKnownValue(
						"podman_network.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("mynetwork"),
					),
				},
			},
			{
				Config: `resource "podman_network" "import_test" {}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"podman_network.import_test",
						tfjsonpath.New("id"),
						knownvalue.StringExact(n.Id),
					),
					statecheck.ExpectKnownValue(
						"podman_network.import_test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(n.Name),
					),
				},
				ResourceName:  "podman_network.import_test",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s@%s", n.Id, framework.Url()),
			},
		},
	})
}

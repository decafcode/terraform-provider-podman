package provider

import (
	"fmt"
	"testing"

	"github.com/decafcode/terraform-provider-podman/internal/api"
	"github.com/decafcode/terraform-provider-podman/internal/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"gotest.tools/v3/assert"
)

func TestAccNetworkResource(t *testing.T) {
	n := &api.NetworkJson{
		Id:   "xyz",
		Name: "importtest",
	}

	apiServer := testutil.ApiServer{
		Networks: []*api.NetworkJson{n},
	}

	framework, err := spawnFramework(t.Context(), &apiServer)
	assert.NilError(t, err)

	defer framework.Stop(t.Context())

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: providerFactories,
		CheckDestroy: func(_ *terraform.State) error {
			if n.Name != "importtest" {
				return apiServer.NetworkWalk(func(n *api.NetworkJson) error {
					return fmt.Errorf("leftover network: %s", n.Name)
				})
			}

			return nil
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
				ImportStateId: fmt.Sprintf("%s,%s", n.Id, framework.Url()),
			},
		},
	})
}

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

func TestAccSecretResource(t *testing.T) {
	s := &api.SecretInspectJson{
		Id: "xyz",
		Spec: api.SecretInspectSpecJson{
			Name: "importtest",
		},
	}

	apiServer := testutil.ApiServer{
		Secrets: []*api.SecretInspectJson{s},
	}

	framework, err := spawnFramework(t.Context(), &apiServer)
	assert.NilError(t, err)

	defer framework.Stop(t.Context())

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: providerFactories,
		CheckDestroy: func(_ *terraform.State) error {
			return apiServer.SecretWalk(func(s *api.SecretInspectJson) error {
				if s.Spec.Name != "importtest" {
					return fmt.Errorf("leftover secret: %s", s.Spec.Name)
				}

				return nil
			})
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "podman_secret" "test" {
						container_host = "%s"
						name           = "mysecret"
						value          = "geheim"
						value_version  = 1
					}
				`, framework.Url()),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"podman_secret.test",
						tfjsonpath.New("id"),
						isString{},
					),
					statecheck.ExpectKnownValue(
						"podman_secret.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("mysecret"),
					),
				},
			},
			{
				Config: `resource "podman_secret" "import_test" {}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"podman_secret.import_test",
						tfjsonpath.New("id"),
						knownvalue.StringExact(s.Id),
					),
					statecheck.ExpectKnownValue(
						"podman_secret.import_test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(s.Spec.Name),
					),
				},
				ResourceName:  "podman_secret.import_test",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", s.Id, framework.Url()),
			},
		},
	})
}

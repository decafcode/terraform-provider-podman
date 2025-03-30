package provider

import (
	"fmt"
	"regexp"
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

func TestAccImageResource(t *testing.T) {
	reference := "example.com/library/hello:v1.0.0"

	apiServer := testutil.ApiServer{
		ValidReferences: map[string]bool{
			reference: true,
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
					resource "podman_image" "test" {
						container_host = "%s"
						policy         = "missing"
						pull_number    = 1
						reference      = "%s"
					}
				`, framework.Url(), reference),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"podman_image.test",
						tfjsonpath.New("id"),
						isString{},
					),
					statecheck.ExpectKnownValue(
						"podman_image.test",
						tfjsonpath.New("policy"),
						knownvalue.StringExact("missing"),
					),
					statecheck.ExpectKnownValue(
						"podman_image.test",
						tfjsonpath.New("pull_number"),
						knownvalue.Int32Exact(1),
					),
					statecheck.ExpectKnownValue(
						"podman_image.test",
						tfjsonpath.New("reference"),
						knownvalue.StringExact(reference),
					),
				},
			},
		},
	})
}

func TestAccImageResourceWithAuth(t *testing.T) {
	auth := api.RegistryAuth{
		Username: "testuser",
		Password: "geheim",
	}

	reference := "example.com/library/hello:v1.0.0"

	apiServer := testutil.ApiServer{
		Auth: &auth,
		ValidReferences: map[string]bool{
			reference: true,
		},
	}

	framework, err := spawnFramework(t.Context(), &apiServer)
	assert.NilError(t, err)

	defer framework.Stop(t.Context())

	env := provider.PodmanProviderEnv{}
	errRegex, err := regexp.Compile("Authentication failed")

	assert.NilError(t, err)

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"podman": providerserver.NewProtocol6WithError(provider.New("test", &env)()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "podman_image" "test_auth" {
						auth = {
							username = "%s"
							password = "%s"
						}

						container_host = "%s"
						policy         = "missing"
						pull_number    = 1
						reference      = "%s"
					}
				`, auth.Username, auth.Password, framework.Url(), reference),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"podman_image.test_auth",
						tfjsonpath.New("id"),
						isString{},
					),
				},
			},
			{
				Config: fmt.Sprintf(`
					resource "podman_image" "test_auth_error" {
						auth = {
							username = "wrong"
							password = "incorrect"
						}

						container_host = "%s"
						policy         = "missing"
						pull_number    = 1
						reference      = "%s"
					}
				`, framework.Url(), reference),
				ExpectError: errRegex,
			},
		},
	})
}

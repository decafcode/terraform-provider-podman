package provider

import (
	"fmt"
	"testing"

	"github.com/decafcode/terraform-provider-podman/internal/api"
	"github.com/decafcode/terraform-provider-podman/internal/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestAccContainerResource(t *testing.T) {
	apiServer := testutil.ApiServer{}
	framework, err := spawnFramework(t.Context(), &apiServer)
	assert.NilError(t, err)

	defer framework.Stop(t.Context())

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: providerFactories,
		CheckDestroy: func(_ *terraform.State) error {
			return apiServer.ContainerWalk(func(c *testutil.TestContainer) error {
				return fmt.Errorf("leftover container: %s", c.Json.Name)
			})
		},
		Steps: []resource.TestStep{
			{
				// Test the various container create JSON fields
				Config: fmt.Sprintf(`
					resource "podman_container" "test" {
						command        = ["cmd"]
						container_host = "%s"
						entrypoint     = ["/bin/sh", "-c"]
						image          = "example.com/library/test:v1.0.0"
						name           = "test"
						restart_policy = "always"
						selinux_options = ["disable"]

						env = {
							"MYENV" = "envvalue"
						}

						labels = {
							"MYLABEL" = "labelvalue"
						}

						mounts = [
							{
								options = ["Z"]
								source = "/srv/data"
								target = "/var/lib/data"
							}
						]

						networks = [
							{
								id = "networkid"
							}
						]

						port_mappings = [
							{
								container_port = 54321
								host_ip        = "192.168.1.1"
								host_port      = 12345
								protocols      = ["tcp", "udp"]
							}
						]

						secrets = [
							{
								gid    = 100
								mode   = parseint("440", 8)
								path   = "/run/secrets/mysecret"
								secret = "secretid"
								uid    = 200
							}
						]

						secret_env = {
							"SECRETENV" = "othersecretid"
						}

						user = {
							group = "mygroup"
							user  = "myuser"
						}
					}
				`, framework.Url()),
				Check: func(_ *terraform.State) error {
					capture, err := apiServer.CaptureContainer("test")

					if err != nil {
						return err
					}

					result := cmp.DeepEqual(capture, &testutil.TestContainer{
						Id: capture.Id,
						Json: api.ContainerCreateJson{
							Command:       []string{"cmd"},
							Entrypoint:    []string{"/bin/sh", "-c"},
							Image:         "example.com/library/test:v1.0.0",
							Name:          "test",
							RestartPolicy: "always",
							SelinuxOpts:   []string{"disable"},

							Env: map[string]string{
								"MYENV": "envvalue",
							},
							Labels: map[string]string{
								"MYLABEL": "labelvalue",
							},
							Mounts: []api.ContainerCreateMountJson{
								{
									Destination: "/var/lib/data",
									Options:     []string{"Z"},
									Source:      "/srv/data",
									Type:        "bind",
								},
							},
							Netns: api.ContainerCreateNamespaceJson{
								NSMode: "bridge",
							},
							Networks: map[string]api.ContainerCreateNetworkJson{
								"networkid": {},
							},
							PortMappings: []api.ContainerCreatePortMappingJson{
								{
									ContainerPort: 54321,
									HostIP:        "192.168.1.1",
									HostPort:      12345,
									Protocol:      "tcp,udp",
								},
							},
							SecretEnv: map[string]string{
								"SECRETENV": "othersecretid",
							},
							Secrets: []api.ContainerCreateSecretJson{
								{
									GID:    100,
									Mode:   0440,
									Source: "secretid",
									Target: "/run/secrets/mysecret",
									UID:    200,
								},
							},
							User: "myuser:mygroup",
						},
						Running: true,
					})()

					if !result.Success() {
						t.Log(result)

						return fmt.Errorf("incorrect post payload")
					}

					return nil
				},
			},
			{
				// Test disabling container auto start
				Config: fmt.Sprintf(`
					resource "podman_container" "test" {
						container_host    = "%s"
						image             = "example.com/library/test:v1.0.0"
						name              = "test"
						start_immediately = false
					}
				`, framework.Url()),
				Check: func(_ *terraform.State) error {
					capture, err := apiServer.CaptureContainer("test")

					if err != nil {
						return err
					}

					if capture.Running {
						return fmt.Errorf("container should not be running")
					}

					return nil
				},
			},
		},
	})
}

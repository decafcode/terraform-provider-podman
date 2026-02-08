package provider

import (
	"encoding/base64"
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

						devices = [
							{
								path = "/dev/dri"
							}
						]

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

						network_namespace = {
							mode = "host"
						}

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
							},
							{
							    container_port = 65432
								host_ip        = "2001:db8::1"
								host_port      = 23456
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

							Devices: []api.ContainerCreateDeviceJson{
								{
									Path: "/dev/dri",
								},
							},
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
								NSMode: "host",
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
								{
									ContainerPort: 65432,
									HostIP:        "2001:db8::1",
									HostPort:      23456,
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
				// Test network namespace defaulting
				Config: fmt.Sprintf(`
					resource "podman_container" "netns" {
						container_host    = "%s"
						image             = "example.com/library/test:v1.0.0"
						name              = "netns"
					}
				`, framework.Url()),
				Check: func(_ *terraform.State) error {
					capture, err := apiServer.CaptureContainer("netns")

					if err != nil {
						return err
					}

					result := cmp.DeepEqual(capture.Json, api.ContainerCreateJson{
						Name:  "netns",
						Image: "example.com/library/test:v1.0.0",
						Netns: api.ContainerCreateNamespaceJson{
							NSMode: "bridge",
						},
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
			{
				// Test create-and-upload, continued in the next step
				Config: fmt.Sprintf(`
					resource "podman_container" "updog" {
						container_host = "%s"
						image          = "example.com/library/test:v1.0.0"
						name           = "updog"

						uploads = [
							{
								content 	 = "Hello"
								content_hash = "hash_of_hello"
								path = "/tmp/file1"
							},
							{
								content 	 = "World"
								content_hash = "hash_of_world"
								path = "/tmp/file2"
							}
						]
					}
				`, framework.Url()),
				Check: func(_ *terraform.State) error {
					capture, err := apiServer.CaptureContainer("updog")

					if err != nil {
						return err
					}

					result := cmp.DeepEqual(capture.UploadLog, []testutil.TestUpload{
						{
							Content: base64.StdEncoding.EncodeToString([]byte("Hello")),
							Mode:    0644,
							Path:    "/tmp/file1",
						},
						{
							Content: base64.StdEncoding.EncodeToString([]byte("World")),
							Mode:    0644,
							Path:    "/tmp/file2",
						},
					})()

					if !result.Success() {
						t.Log(result)

						return fmt.Errorf("upload log is incorrect (part 1)")
					}

					return nil
				},
			},
			{
				// Test in-place re-upload to an existing container.
				//
				// Note the change to the first but not the second entry
				// in the uploads array. Only this changed file should get
				// uploaded, and the container itself must be preserved
				// between steps (this is indirectly tested by the presence of
				// the old file contents in the upload log, which is not as
				// rigorous as I'd like but oh well).
				Config: fmt.Sprintf(`
					resource "podman_container" "updog" {
						container_host = "%s"
						image          = "example.com/library/test:v1.0.0"
						name           = "updog"

						uploads = [
							{
								content 	 = "Howdy"
								content_hash = "hash_of_howdy"
								path = "/tmp/file1"
							},
							{
								content 	 = "World"
								content_hash = "hash_of_world"
								path = "/tmp/file2"
							}
						]
					}
				`, framework.Url()),
				Check: func(_ *terraform.State) error {
					capture, err := apiServer.CaptureContainer("updog")

					if err != nil {
						return err
					}

					result := cmp.DeepEqual(capture.UploadLog, []testutil.TestUpload{
						{
							Content: base64.StdEncoding.EncodeToString([]byte("Hello")),
							Mode:    0644,
							Path:    "/tmp/file1",
						},
						{
							Content: base64.StdEncoding.EncodeToString([]byte("World")),
							Mode:    0644,
							Path:    "/tmp/file2",
						},
						{
							Content:    base64.StdEncoding.EncodeToString([]byte("Howdy")),
							Mode:       0644,
							Path:       "/tmp/file1",
							WasRunning: true,
						},
					})()

					if !result.Success() {
						t.Log(result)

						return fmt.Errorf("upload log is incorrect (part 2)")
					}

					return nil
				},
			},
		},
	})
}

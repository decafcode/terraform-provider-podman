package client

import (
	"testing"

	"github.com/decafcode/terraform-provider-podman/internal/api"
	"github.com/decafcode/terraform-provider-podman/internal/testutil"
	"gotest.tools/v3/assert"
)

func testImagePullSuccess(t *testing.T, auth *api.RegistryAuth) {
	reference := "example.com/foo/bar:v1"
	policy := "always"
	apiServer := &testutil.ApiServer{
		Auth: auth,
		ValidReferences: map[string]bool{
			reference: true,
		},
	}

	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)
	defer f.Stop(t.Context())

	ch, err := f.ImagePull(t.Context(), api.ImagePullQuery{
		Reference: reference,
		Policy:    policy,
	}, auth)

	assert.NilError(t, err)

	var gotProgress bool
	var id string

	for event := range ch {
		errEvent, match := event.(api.ImagePullErrorEvent)

		if match {
			t.Fatal(errEvent)
		}

		idEvent, match := event.(api.ImagePullImagesEvent)

		if match {
			id = idEvent.Id
		}

		_, match = event.(api.ImagePullStreamEvent)

		if match {
			// Any reasonable implementation of the Podman API will send at
			// least one progress message during a successful image pull.

			gotProgress = true
		}

		t.Logf("%#v\n", event)
	}

	assert.Assert(t, id != "")
	assert.Assert(t, gotProgress)

	apiServer.Mutex.Lock()
	defer apiServer.Mutex.Unlock()

	assert.DeepEqual(t, apiServer.PullRequests[0], testutil.PullRequest{
		Reference: reference,
		Policy:    policy,
	})
}

func TestImagePullAnon(t *testing.T) {
	testImagePullSuccess(t, nil)
}

func TestImagePullAuth(t *testing.T) {
	testImagePullSuccess(t, &api.RegistryAuth{
		Username: "user",
		Password: "pass",
	})
}

func TestImagePullFailure(t *testing.T) {
	apiServer := &testutil.ApiServer{}
	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)
	defer f.Stop(t.Context())

	ch, err := f.ImagePull(t.Context(), api.ImagePullQuery{
		Reference: "invalidref",
	}, nil)

	assert.NilError(t, err)

	var gotError bool

	for event := range ch {
		_, match := event.(api.ImagePullErrorEvent)

		if match {
			gotError = true
		}

		t.Logf("%#v\n", event)
	}

	assert.Assert(t, gotError)
}

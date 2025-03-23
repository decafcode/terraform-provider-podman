package client

import (
	"testing"

	"github.com/decafcode/terraform-provider-podman/internal/api"
	"github.com/decafcode/terraform-provider-podman/internal/testutil"
	"gotest.tools/v3/assert"
	testCmp "gotest.tools/v3/assert/cmp"
)

func TestSecretCreate(t *testing.T) {
	apiServer := &testutil.ApiServer{}

	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)

	defer f.Stop(t.Context())

	name, value := "test", "geheim"
	result, err := f.SecretCreate(t.Context(), name, value)
	assert.NilError(t, err)

	if result.Id == "" {
		t.Error("no id")
	}

	apiServer.Mutex.Lock()
	defer apiServer.Mutex.Unlock()

	stored := apiServer.Secrets[result.Id]
	assert.Assert(t, testCmp.Equal(name, stored.Spec.Name))
	assert.Assert(t, testCmp.Equal(value, stored.SecretData))
}

func TestSecretGet(t *testing.T) {
	s1 := &api.SecretInspectJson{
		Id: "1",
		Spec: api.SecretInspectSpecJson{
			Name: "one",
		},
	}

	s2 := &api.SecretInspectJson{
		Id: "2",
		Spec: api.SecretInspectSpecJson{
			Name: "two",
		},
	}

	apiServer := &testutil.ApiServer{
		Secrets: map[string]*api.SecretInspectJson{
			s1.Id: s1,
			s2.Id: s2,
		},
	}

	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)

	defer f.Stop(t.Context())

	actual, err := f.SecretInspect(t.Context(), s2.Id)
	assert.NilError(t, err)
	assert.DeepEqual(t, s2, actual)
}

func TestSecretDelete(t *testing.T) {
	s := &api.SecretInspectJson{
		Id: "1234",
		Spec: api.SecretInspectSpecJson{
			Name: "one",
		},
	}

	apiServer := &testutil.ApiServer{
		Secrets: map[string]*api.SecretInspectJson{
			s.Id: s,
		},
	}

	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)

	defer f.Stop(t.Context())

	err = f.SecretDelete(t.Context(), s.Id)
	assert.NilError(t, err)

	apiServer.Mutex.Lock()
	defer apiServer.Mutex.Unlock()

	if len(apiServer.Secrets) > 0 {
		t.Error("secret not deleted")
	}
}

package client

import (
	"testing"

	"github.com/decafcode/terraform-provider-podman/internal/api"
	"github.com/decafcode/terraform-provider-podman/internal/testutil"
	"gotest.tools/v3/assert"
	testCmp "gotest.tools/v3/assert/cmp"
)

func TestNetworkCreate(t *testing.T) {
	apiServer := &testutil.ApiServer{}

	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)

	defer f.Stop(t.Context())

	network := &api.NetworkJson{
		Name: "test",
	}

	result, err := f.NetworkCreate(t.Context(), network)
	assert.NilError(t, err)
	assert.Assert(t, testCmp.Equal(network.Name, result.Name))

	if result.Id == "" {
		t.Error("no id")
	}

	apiServer.Mutex.Lock()
	defer apiServer.Mutex.Unlock()

	stored := apiServer.Networks[result.Id]
	assert.Assert(t, testCmp.Equal(network.Name, stored.Name))
}

func TestNetworkGet(t *testing.T) {
	n1 := &api.NetworkJson{
		Id:   "1",
		Name: "one",
	}

	n2 := &api.NetworkJson{
		Id:   "2",
		Name: "two",
	}

	apiServer := &testutil.ApiServer{
		Networks: map[string]*api.NetworkJson{
			n1.Id: n1,
			n2.Id: n2,
		},
	}

	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)

	defer f.Stop(t.Context())

	actual, err := f.NetworkInspect(t.Context(), n2.Id)
	assert.NilError(t, err)
	assert.DeepEqual(t, n2, actual)
}

func TestNetworkDelete(t *testing.T) {
	n := &api.NetworkJson{
		Id:   "1234",
		Name: "one/two&three four",
	}

	apiServer := &testutil.ApiServer{
		Networks: map[string]*api.NetworkJson{
			n.Id: n,
		},
	}

	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)

	defer f.Stop(t.Context())

	err = f.NetworkDelete(t.Context(), n.Id)
	assert.NilError(t, err)

	apiServer.Mutex.Lock()
	defer apiServer.Mutex.Unlock()

	if len(apiServer.Networks) > 0 {
		t.Error("network not deleted")
	}
}

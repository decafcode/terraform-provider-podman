package client

import (
	"testing"

	"github.com/decafcode/terraform-provider-podman/internal/api"
	"github.com/decafcode/terraform-provider-podman/internal/testutil"
	"gotest.tools/v3/assert"
)

func TestContainerCreate(t *testing.T) {
	apiServer := &testutil.ApiServer{}

	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)

	defer f.Stop(t.Context())

	container := &api.ContainerCreateJson{
		Image: "example.com/library/test:v1.0.0",
		Name:  "test",
	}

	result, err := f.ContainerCreate(t.Context(), container)
	assert.NilError(t, err)

	if result.Id == "" {
		t.Error("no id")
	}
}

func TestContainerDelete(t *testing.T) {
	c := &testutil.TestContainer{
		Id:   "1",
		Json: api.ContainerCreateJson{Name: "one"},
	}

	apiServer := &testutil.ApiServer{
		Containers: []*testutil.TestContainer{c},
	}

	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)

	defer f.Stop(t.Context())

	err = f.ContainerDelete(t.Context(), c.Json.Name)
	assert.NilError(t, err)

	_, err = f.ContainerInspect(t.Context(), c.Json.Name)
	assert.ErrorContains(t, err, "not found")
}

func TestContainerInspect(t *testing.T) {
	c1 := &testutil.TestContainer{
		Id:   "1",
		Json: api.ContainerCreateJson{Name: "one"},
	}

	c2 := &testutil.TestContainer{
		Id:   "2",
		Json: api.ContainerCreateJson{Name: "two"},
	}

	apiServer := &testutil.ApiServer{
		Containers: []*testutil.TestContainer{c1, c2},
	}

	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)

	defer f.Stop(t.Context())

	actual, err := f.ContainerInspect(t.Context(), c2.Json.Name)
	assert.NilError(t, err)
	assert.Equal(t, c2.Json.Name, actual.Name)
}

func TestContainerRename(t *testing.T) {
	c := &testutil.TestContainer{
		Id:   "1",
		Json: api.ContainerCreateJson{Name: "before"},
	}

	apiServer := &testutil.ApiServer{
		Containers: []*testutil.TestContainer{c},
	}

	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)

	defer f.Stop(t.Context())

	err = f.ContainerRename(t.Context(), c.Id, "after")
	assert.NilError(t, err)

	result, err := f.ContainerInspect(t.Context(), c.Json.Name)
	assert.NilError(t, err)
	assert.Equal(t, result.Name, "after")
}

func TestContainerStart(t *testing.T) {
	c := &testutil.TestContainer{
		Id:      "1",
		Json:    api.ContainerCreateJson{Name: "one"},
		Running: false,
	}

	apiServer := &testutil.ApiServer{
		Containers: []*testutil.TestContainer{c},
	}

	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)

	defer f.Stop(t.Context())

	err = f.ContainerStart(t.Context(), c.Id)
	assert.NilError(t, err)

	result, err := apiServer.CaptureContainer(c.Json.Name)
	assert.NilError(t, err)
	assert.Equal(t, result.Running, true)
}

func TestContainerStop(t *testing.T) {
	c := &testutil.TestContainer{
		Id:      "1",
		Json:    api.ContainerCreateJson{Name: "one"},
		Running: true,
	}

	apiServer := &testutil.ApiServer{
		Containers: []*testutil.TestContainer{c},
	}

	f, err := spawnFramework(t.Context(), apiServer)
	assert.NilError(t, err)

	defer f.Stop(t.Context())

	err = f.ContainerStop(t.Context(), c.Id)
	assert.NilError(t, err)

	result, err := apiServer.CaptureContainer(c.Json.Name)
	assert.NilError(t, err)
	assert.Equal(t, result.Running, false)
}

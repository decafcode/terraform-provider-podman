package testutil

import (
	"archive/tar"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/decafcode/terraform-provider-podman/internal/api"
)

func (s *ApiServer) lookupContainer(nameOrId string) (*TestContainer, error) {
	for _, c := range s.Containers {
		if c.Json.Name == nameOrId || c.Id == nameOrId {
			return c, nil
		}
	}

	return nil, statusError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("nameOrIf \"%s\" not found", nameOrId),
	}
}

func (s *ApiServer) handleContainerArchive(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	container, err := s.lookupContainer(nameOrId)

	if err != nil {
		return err
	}

	if req.Header.Get("content-type") != "application/x-tar" {
		return statusError{
			Code:    http.StatusUnsupportedMediaType,
			Message: "Expected application/x-tar",
		}
	}

	reader := tar.NewReader(req.Body)

	for {
		header, err := reader.Next()

		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				return err
			}
		}

		bytes := make([]byte, header.Size)
		n, err := reader.Read(bytes)

		if err != nil && err != io.EOF {
			return err
		}

		if int64(n) != header.Size {
			return fmt.Errorf("truncated read for %s", header.Name)
		}

		container.UploadLog = append(container.UploadLog, TestUpload{
			Content:    base64.StdEncoding.EncodeToString(bytes),
			Gid:        header.Gid,
			Mode:       header.Mode,
			Path:       header.Name,
			WasRunning: container.Running,
			Uid:        header.Uid,
		})
	}
}

func (s *ApiServer) handleContainerCreate(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	c := &TestContainer{}
	err := readJson(req, &c.Json)

	if err != nil {
		return err
	}

	s.nextId++
	c.Id = fmt.Sprintf("%d", s.nextId)
	s.Containers = append(s.Containers, c)

	result := &api.ContainerCreatedJson{Id: c.Id}

	return writeJson(resp, result)
}

func (s *ApiServer) handleContainerDelete(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	_, err := s.lookupContainer(nameOrId)

	if err != nil {
		return nil
	}

	s.Containers = slices.DeleteFunc(s.Containers, func(c *TestContainer) bool {
		return c.Json.Name == nameOrId || c.Id == nameOrId
	})

	return nil
}

func (s *ApiServer) handleContainerGet(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	match, err := s.lookupContainer(nameOrId)

	if err != nil {
		return err
	}

	result := api.ContainerInspectJson{Name: match.Json.Name}

	return writeJson(resp, result)
}

func (s *ApiServer) handleContainerRename(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	query := req.URL.Query()
	newName := query.Get("name")

	if newName == "" {
		return statusError{
			Code:    http.StatusBadRequest,
			Message: "name query parameter is missing or empty",
		}
	}

	nameOrId := req.PathValue("nameOrId")
	match, err := s.lookupContainer(nameOrId)

	if err != nil {
		return err
	}

	match.Json.Name = newName

	return nil
}

func (s *ApiServer) handleContainerStart(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	match, err := s.lookupContainer(nameOrId)

	if err != nil {
		return err
	}

	if match.Running {
		resp.WriteHeader(http.StatusNotModified)
	}

	match.Running = true

	return nil
}

func (s *ApiServer) handleContainerStop(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	match, err := s.lookupContainer(nameOrId)

	if err != nil {
		return err
	}

	query := req.URL.Query()

	if query.Get("ignore") != "true" && !match.Running {
		resp.WriteHeader(http.StatusNotModified)
	}

	match.Running = false

	return nil
}

// Snapshot the internal state of a container spec inside the test API server
// and return a copy. This mostly adheres to the JSON format of a container
// create request rather than using the response format of a container inspect
// request, which is completely and pointlessly different and also isn't a
// message format that this provider cares about beyond using it to verify the
// current name of a container resource.
//
// It also gives access to a log of uploads to the container's filesystem.
func (s *ApiServer) CaptureContainer(nameOrId string) (*TestContainer, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	match, err := s.lookupContainer(nameOrId)

	if err != nil {
		return nil, err
	}

	return match.Clone(), nil
}

func (s *ApiServer) ContainerWalk(callback func(c *TestContainer) error) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, c := range s.Containers {
		err := callback(c)

		if err != nil {
			return err
		}
	}

	return nil
}

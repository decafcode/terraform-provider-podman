package testutil

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/decafcode/terraform-provider-podman/internal/api"
)

func (s *ApiServer) lookupNetwork(nameOrId string) (*api.NetworkJson, error) {
	for _, n := range s.Networks {
		if n.Id == nameOrId || n.Name == nameOrId {
			return n, nil
		}
	}

	return nil, statusError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("nameOrId \"%s\" not found", nameOrId),
	}
}

func (s *ApiServer) handleNetworkCreate(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var json *api.NetworkJson
	err := readJson(req, &json)

	if err != nil {
		return err
	}

	if json.Id != "" {
		return statusError{
			Code:    http.StatusBadRequest,
			Message: "id must not be present",
		}
	}

	s.nextId++
	n := *json
	n.Id = fmt.Sprintf("%d", s.nextId)
	s.Networks = append(s.Networks, &n)

	return writeJson(resp, n)
}

func (s *ApiServer) handleNetworkDelete(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	_, err := s.lookupNetwork(nameOrId)

	if err != nil {
		return err
	}

	s.Networks = slices.DeleteFunc(s.Networks, func(n *api.NetworkJson) bool {
		return n.Name == nameOrId || n.Id == nameOrId
	})

	return nil
}

func (s *ApiServer) handleNetworkGet(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	match, err := s.lookupNetwork(nameOrId)

	if err != nil {
		return err
	}

	return writeJson(resp, match)
}

func (s *ApiServer) NetworkWalk(callback func(c *api.NetworkJson) error) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, n := range s.Networks {
		err := callback(n)

		if err != nil {
			return err
		}
	}

	return nil
}

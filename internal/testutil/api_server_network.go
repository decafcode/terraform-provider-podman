package testutil

import (
	"context"
	"fmt"
	"net/http"

	"github.com/decafcode/terraform-provider-podman/internal/api"
)

func (s *ApiServer) lookupNetwork(nameOrId string) (*api.NetworkJson, error) {
	match := s.Networks[nameOrId]

	if match == nil {
		return nil, statusError{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("ID \"%s\" not found, lookup by name not implemented.", nameOrId),
		}
	}

	return match, nil
}

func (s *ApiServer) handleNetworkCreate(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

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

	if s.Networks == nil {
		s.Networks = make(map[string]*api.NetworkJson)
	}

	s.nextId++
	dupe := *json
	dupe.Id = fmt.Sprintf("%d", s.nextId)
	s.Networks[dupe.Id] = &dupe

	return writeJson(resp, dupe)
}

func (s *ApiServer) handleNetworkDelete(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	_, err := s.lookupNetwork(nameOrId)

	if err != nil {
		return err
	}

	delete(s.Networks, nameOrId)

	return nil
}

func (s *ApiServer) handleNetworkGet(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	match, err := s.lookupNetwork(nameOrId)

	if err != nil {
		return err
	}

	return writeJson(resp, match)
}

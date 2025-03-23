package testutil

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/decafcode/terraform-provider-podman/internal/api"
)

func (s *ApiServer) lookupSecret(nameOrId string) (*api.SecretInspectJson, error) {
	match := s.Secrets[nameOrId]

	if match == nil {
		return nil, statusError{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("ID \"%s\" not found, lookup by name not implemented.", nameOrId),
		}
	}

	return match, nil
}

func (s *ApiServer) handleSecretCreate(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	query := req.URL.Query()

	if !query.Has("name") {
		return statusError{
			Code:    http.StatusBadRequest,
			Message: "No name query param",
		}
	}

	name := query.Get("name")
	bytes, err := io.ReadAll(req.Body)

	if err != nil {
		return err
	}

	if s.Secrets == nil {
		s.Secrets = make(map[string]*api.SecretInspectJson)
	}

	existing := s.Secrets[name]

	if existing != nil {
		return statusError{
			Code:    http.StatusConflict,
			Message: "name conflict",
		}
	}

	s.nextId++
	storedSecret := &api.SecretInspectJson{
		Id:         fmt.Sprintf("%d", s.nextId),
		SecretData: string(bytes),
		Spec: api.SecretInspectSpecJson{
			Name: name,
		},
	}

	s.Secrets[storedSecret.Id] = storedSecret

	out := &api.SecretCreateJson{
		Id: storedSecret.Id,
	}

	return writeJson(resp, out)
}

func (s *ApiServer) handleSecretDelete(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	_, err := s.lookupSecret(nameOrId)

	if err != nil {
		return err
	}

	delete(s.Secrets, nameOrId)

	return nil
}

func (s *ApiServer) handleSecretGet(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	match, err := s.lookupSecret(nameOrId)

	if err != nil {
		return err
	}

	return writeJson(resp, match)
}

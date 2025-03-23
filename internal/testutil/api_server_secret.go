package testutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/decafcode/terraform-provider-podman/internal/api"
)

func (s *ApiServer) lookupSecret(nameOrId string) (*api.SecretInspectJson, error) {
	for _, secret := range s.Secrets {
		if secret.Spec.Name == nameOrId || secret.Id == nameOrId {
			return secret, nil
		}
	}

	return nil, statusError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("nameOrId \"%s\" not found", nameOrId),
	}
}

func (s *ApiServer) handleSecretCreate(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

	for _, secret := range s.Secrets {
		if secret.Spec.Name == name {
			return statusError{
				Code:    http.StatusConflict,
				Message: "name conflict",
			}
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

	s.Secrets = append(s.Secrets, storedSecret)

	out := &api.SecretCreateJson{
		Id: storedSecret.Id,
	}

	return writeJson(resp, out)
}

func (s *ApiServer) handleSecretDelete(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	_, err := s.lookupSecret(nameOrId)

	if err != nil {
		return err
	}

	s.Secrets = slices.DeleteFunc(s.Secrets, func(s *api.SecretInspectJson) bool {
		return s.Spec.Name == nameOrId || s.Id == nameOrId
	})

	return nil
}

func (s *ApiServer) handleSecretGet(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	match, err := s.lookupSecret(nameOrId)

	if err != nil {
		return err
	}

	return writeJson(resp, match)
}

func (s *ApiServer) SecretWalk(callback func(c *api.SecretInspectJson) error) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, secret := range s.Secrets {
		err := callback(secret)

		if err != nil {
			return err
		}
	}

	return nil
}

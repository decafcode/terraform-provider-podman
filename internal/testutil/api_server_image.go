package testutil

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/decafcode/terraform-provider-podman/internal/api"
)

var newline = []byte("\n")

func writeEvent(resp http.ResponseWriter, obj any) {
	bytes, err := json.Marshal(obj)

	if err != nil {
		fmt.Printf("Error serializing event: %v\n", err)

		return
	}

	withNewline := slices.Concat(bytes, newline)
	_, err = resp.Write(withNewline)

	if err != nil {
		fmt.Printf("Error transmitting event: %v\n", err)

		return
	}

	flusher, ok := resp.(http.Flusher)

	if ok {
		flusher.Flush()
	}
}

func (s *ApiServer) lookupImage(nameOrId string) (*api.ImageJson, error) {
	for _, img := range s.Images {
		if img.Id == nameOrId || slices.Contains(img.Names, nameOrId) {
			return img, nil
		}
	}

	return nil, statusError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("nameOrId \"%s\" not found", nameOrId),
	}
}

func (s *ApiServer) handleImageDelete(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	_, err := s.lookupImage(nameOrId)

	if err != nil {
		return err
	}

	s.Images = slices.DeleteFunc(s.Images, func(img *api.ImageJson) bool {
		return img.Id == nameOrId || slices.Contains(img.Names, nameOrId)
	})

	return nil
}

func (s *ApiServer) handleImageGet(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	nameOrId := req.PathValue("nameOrId")
	match, err := s.lookupImage(nameOrId)

	if err != nil {
		return err
	}

	return writeJson(resp, match)
}

func (s *ApiServer) handleImagePull(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	values := req.URL.Query()
	policy := values.Get("policy")
	reference := values.Get("reference")

	if reference == "" {
		return statusError{
			Code:    http.StatusBadRequest,
			Message: "Reference parameter is missing",
		}
	}

	resp.Header().Add("content-type", "application/octet-stream")
	resp.WriteHeader(http.StatusOK)

	if !s.ValidReferences[reference] {
		writeEvent(resp, api.ImagePullErrorEvent{
			Error: fmt.Sprintf("Not present in valid references list: %s", reference),
		})

		return nil
	}

	if s.Auth != nil {
		authHeader := req.Header.Get("x-registry-auth")

		if authHeader == "" {
			writeEvent(resp, api.ImagePullErrorEvent{
				Error: "Authentication required",
			})

			return nil
		}

		authJson, err := base64.URLEncoding.DecodeString(authHeader)

		if err != nil {
			writeEvent(resp, api.ImagePullErrorEvent{
				Error: err.Error(),
			})

			return nil
		}

		var auth api.RegistryAuth
		err = json.Unmarshal(authJson, &auth)

		if err != nil {
			writeEvent(resp, api.ImagePullErrorEvent{
				Error: err.Error(),
			})

			return nil
		}

		// Comparison is not timing safe, but this is a test harness so we don't care.
		if auth.Username != s.Auth.Username || auth.Password != s.Auth.Password {
			writeEvent(resp, api.ImagePullErrorEvent{
				Error: "Authentication failed",
			})

			return nil
		}
	}

	s.nextId++
	idStr := fmt.Sprintf("%d", s.nextId)

	s.PullRequests = append(s.PullRequests, PullRequest{
		Policy:    policy,
		Reference: reference,
	})

	json := &api.ImageJson{
		Id:    idStr,
		Names: []string{reference},
	}

	s.Images = append(s.Images, json)

	writeEvent(resp, api.ImagePullStreamEvent{
		Stream: "Progress message goes here",
	})

	writeEvent(resp, api.ImagePullImagesEvent{
		Id:     idStr,
		Images: []string{reference},
	})

	return nil
}

func (s *ApiServer) ImageWalk(callback func(c *api.ImageJson) error) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, img := range s.Images {
		err := callback(img)

		if err != nil {
			return err
		}
	}

	return nil
}

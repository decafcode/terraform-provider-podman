package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/decafcode/terraform-provider-podman/internal/api"
)

type PullRequest struct {
	Reference string
	Policy    string
}

type ApiServer struct {
	Auth            *api.RegistryAuth
	Images          map[string]*api.ImageJson
	Networks        map[string]*api.NetworkJson
	Mutex           sync.Mutex
	PullRequests    []PullRequest
	Secrets         map[string]*api.SecretInspectJson
	ValidReferences map[string]bool

	nextId int
}

func writeJson(resp http.ResponseWriter, v any) error {
	resp.Header().Add("content-type", "application/json")

	return json.NewEncoder(resp).Encode(v)
}

func readJson(req *http.Request, v any) error {
	contentType := req.Header.Get("content-type")

	if contentType != "application/json" {
		return statusError{
			Code:    http.StatusUnsupportedMediaType,
			Message: fmt.Sprintf("unexpected content type %s", contentType),
		}
	}

	return json.NewDecoder(req.Body).Decode(v)
}

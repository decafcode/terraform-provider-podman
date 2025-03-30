package client

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/decafcode/terraform-provider-podman/internal/api"
)

func (c *Client) ImageDelete(ctx context.Context, nameOrId string) error {
	relPath := fmt.Sprintf("v5.0.0/libpod/images/%s", url.PathEscape(nameOrId))

	return c.resourceDelete(ctx, relPath)
}

func (c *Client) ImagePull(ctx context.Context, query api.ImagePullQuery, auth *api.RegistryAuth) (<-chan any, error) {
	path, err := url.Parse("v5.0.0/libpod/images/pull")

	if err != nil {
		panic(err)
	}

	values := make(url.Values)
	values.Add("policy", query.Policy)
	values.Add("reference", query.Reference)
	path.RawQuery = values.Encode()

	url := c.urlBase.ResolveReference(path).String()
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)

	if err != nil {
		return nil, err
	}

	if auth != nil {
		authBytes, err := json.Marshal(auth)

		if err != nil {
			return nil, err
		}

		req.Header.Add("x-registry-auth", base64.StdEncoding.EncodeToString(authBytes))
	}

	resp, err := c.http.Do(req)

	if err != nil {
		return nil, err
	}

	err = checkStatus(resp)

	if err != nil {
		return nil, err
	}

	result := make(chan any)

	go streamEvents(resp.Body, result)

	return result, nil
}

func streamEvents(in io.ReadCloser, out chan<- any) {
	defer close(out)
	defer in.Close()

	scanner := bufio.NewScanner(in)

	for scanner.Scan() {
		var obj map[string]any
		line := scanner.Bytes()
		err := json.Unmarshal(line, &obj)

		if err != nil {
			out <- err

			return
		}

		if obj["error"] != nil {
			var item api.ImagePullErrorEvent
			err := json.Unmarshal(line, &item)

			if err != nil {
				out <- err

				return
			}

			out <- item
		}

		if obj["images"] != nil {
			var item api.ImagePullImagesEvent
			err := json.Unmarshal(line, &item)

			if err != nil {
				out <- err

				return
			}

			out <- item
		}

		if obj["stream"] != nil {
			var item api.ImagePullStreamEvent
			err := json.Unmarshal(line, &item)

			if err != nil {
				out <- err

				return
			}

			out <- item
		}
	}

	err := scanner.Err()

	if err != nil {
		out <- err
	}
}

package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/decafcode/terraform-provider-podman/internal/api"
)

func (c *Client) SecretCreate(ctx context.Context, name, value string) (*api.SecretCreateJson, error) {
	relUrl, err := url.Parse("v5.0.0/libpod/secrets/create")

	if err != nil {
		panic(err)
	}

	params := make(url.Values)
	params.Add("name", name)
	relUrl.RawQuery = params.Encode()

	vb := []byte(value)
	absUrl := c.urlBase.ResolveReference(relUrl).String()
	req, err := http.NewRequestWithContext(ctx, "POST", absUrl, bytes.NewReader(vb))

	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "text/plain;charset=UTF-8")
	resp, err := c.http.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var out *api.SecretCreateJson
	err = readJson(resp, &out)

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *Client) SecretDelete(ctx context.Context, nameOrId string) error {
	path := fmt.Sprintf("v5.0.0/libpod/secrets/%s", url.PathEscape(nameOrId))

	return c.resourceDelete(ctx, path)
}

func (c *Client) SecretInspect(ctx context.Context, nameOrId string) (*api.SecretInspectJson, error) {
	var out *api.SecretInspectJson
	path := fmt.Sprintf("v5.0.0/libpod/secrets/%s/json", url.PathEscape(nameOrId))
	err := c.resourceGet(ctx, path, &out)

	if err != nil {
		return nil, err
	}

	return out, nil
}

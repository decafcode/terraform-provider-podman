package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/decafcode/terraform-provider-podman/internal/api"
)

func (c *Client) NetworkCreate(ctx context.Context, in *api.NetworkJson) (*api.NetworkJson, error) {
	var out *api.NetworkJson
	err := c.resourceCreate(ctx, "v5.0.0/libpod/networks/create", in, &out)

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *Client) NetworkDelete(ctx context.Context, nameOrId string) error {
	path := fmt.Sprintf("v5.0.0/libpod/networks/%s", url.PathEscape(nameOrId))

	return c.resourceDelete(ctx, path)
}

func (c *Client) NetworkInspect(ctx context.Context, nameOrId string) (*api.NetworkJson, error) {
	var out *api.NetworkJson
	path := fmt.Sprintf("v5.0.0/libpod/networks/%s/json", url.PathEscape(nameOrId))
	err := c.resourceGet(ctx, path, &out)

	if err != nil {
		return nil, err
	}

	return out, nil
}

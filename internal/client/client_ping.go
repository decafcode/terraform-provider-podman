package client

import (
	"context"
	"net/http"
	"net/url"
)

func (c *Client) Ping(ctx context.Context) error {
	path, err := url.Parse("v5.0.0/libpod/_ping")

	if err != nil {
		panic(err)
	}

	url := c.urlBase.ResolveReference(path).String()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return checkStatus(resp)
}

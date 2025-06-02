package client

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/decafcode/terraform-provider-podman/internal/api"
)

type archiveBuilder func(writer *tar.Writer) error

func (c *Client) sendArchiveTask(ctx context.Context, nameOrId string, reader *io.PipeReader, promise chan<- error) {
	path := fmt.Sprintf("v5.0.0/libpod/containers/%s/archive?path=%%2F", url.PathEscape(nameOrId))
	promise <- c.resourceStream(ctx, path, "application/x-tar", reader)

	close(promise)
}

func (c *Client) ContainerArchive(ctx context.Context, nameOrId string, builder archiveBuilder) error {
	reader, writer := io.Pipe()
	promise := make(chan error)

	go c.sendArchiveTask(ctx, nameOrId, reader, promise)

	arc := tar.NewWriter(writer)
	err := builder(arc)

	if err == nil { // equal, not not-equal!
		err = arc.Close()
	}

	writer.CloseWithError(err)

	if err != nil {
		return err
	}

	return <-promise
}

func (c *Client) ContainerCreate(ctx context.Context, in *api.ContainerCreateJson) (*api.ContainerCreatedJson, error) {
	var out *api.ContainerCreatedJson
	err := c.resourceCreate(ctx, "v5.0.0/libpod/containers/create", in, &out)

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *Client) ContainerDelete(ctx context.Context, nameOrId string) error {
	path := fmt.Sprintf("v5.0.0/libpod/containers/%s", url.PathEscape(nameOrId))

	return c.resourceDelete(ctx, path)
}

func (c *Client) ContainerInspect(ctx context.Context, nameOrId string) (*api.ContainerInspectJson, error) {
	var out *api.ContainerInspectJson
	path := fmt.Sprintf("v5.0.0/libpod/containers/%s/json", url.PathEscape(nameOrId))
	err := c.resourceGet(ctx, path, &out)

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *Client) ContainerRename(ctx context.Context, nameOrId, newName string) error {
	path := fmt.Sprintf(
		"v5.0.0/libpod/containers/%s/rename?name=%s",
		url.PathEscape(nameOrId),
		url.PathEscape(newName))

	return c.resourceSignal(ctx, path)
}

func (c *Client) ContainerStart(ctx context.Context, nameOrId string) error {
	path := fmt.Sprintf("v5.0.0/libpod/containers/%s/start", url.PathEscape(nameOrId))

	return c.resourceSignal(ctx, path)
}

func (c *Client) ContainerStop(ctx context.Context, nameOrId string) error {
	path := fmt.Sprintf("v5.0.0/libpod/containers/%s/stop?ignore=true", url.PathEscape(nameOrId))
	err := c.resourceSignal(ctx, path)

	if err != nil {
		status, ok := err.(StatusCodeError)

		if ok && status.StatusCode == http.StatusNotModified {
			return nil
		}
	}

	return err
}

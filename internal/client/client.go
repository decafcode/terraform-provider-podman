package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	transport io.Closer
	http      *http.Client
	urlBase   *url.URL
}

type Config struct {
	Ssh ssh.ClientConfig
}

func Connect(ctx context.Context, url *url.URL, config *Config) (*Client, error) {
	switch url.Scheme {
	case "tcp":
		urlCopy := *url
		urlCopy.Scheme = "http"

		return &Client{
			http:    &http.Client{},
			urlBase: &urlCopy,
		}, nil

	case "ssh":
		transport, http, err := dialSshTransport(ctx, url, &config.Ssh)

		if err != nil {
			return nil, err
		}

		dummyBase, err := url.Parse("http://UNIX-OVER-SSH/")

		if err != nil {
			panic(err)
		}

		return &Client{
			transport: transport,
			http:      http,
			urlBase:   dummyBase,
		}, nil

	case "unix":
		http := createUnixTransport(url.Path)
		dummyBase, err := url.Parse("http://LOCAL-UNIX-SOCKET/")

		if err != nil {
			panic(err)
		}

		return &Client{
			http:    http,
			urlBase: dummyBase,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported URL scheme: %s", url.Scheme)
	}
}

func (c *Client) Close() error {
	c.http.CloseIdleConnections()

	if c.transport != nil {
		err := c.transport.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func checkStatus(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var message string
		msgBytes, err := io.ReadAll(resp.Body)

		if err != nil {
			message = fmt.Sprintf("(error reading response body: %v)", err)
		} else {
			message = string(msgBytes)
		}

		return StatusCodeError{
			HttpError: HttpError{
				URL: resp.Request.URL,
			},
			StatusCode: resp.StatusCode,
			Message:    message,
		}
	}

	return nil
}

func readJson(resp *http.Response, out any) error {
	err := checkStatus(resp)

	if err != nil {
		return err
	}

	contentType := resp.Header.Get("content-type")

	if contentType != "application/json" {
		return ContentTypeError{
			HttpError: HttpError{
				URL: resp.Request.URL,
			},
			ContentType: contentType,
		}
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func pipeJson(writer *io.PipeWriter, v any) {
	err := json.NewEncoder(writer).Encode(v)
	writer.CloseWithError(err)
}

func (c *Client) resourceCreate(ctx context.Context, path string, in any, out any) error {
	relUrl, err := url.Parse(path)

	if err != nil {
		return err
	}

	reader, writer := io.Pipe()

	go pipeJson(writer, in)

	absUrl := c.urlBase.ResolveReference(relUrl).String()
	req, err := http.NewRequestWithContext(ctx, "POST", absUrl, reader)

	if err != nil {
		return err
	}

	req.Header.Add("content-type", "application/json")
	resp, err := c.http.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	err = readJson(resp, &out)

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) resourceDelete(ctx context.Context, path string) error {
	relUrl, err := url.Parse(path)

	if err != nil {
		return err
	}

	absUrl := c.urlBase.ResolveReference(relUrl).String()
	req, err := http.NewRequestWithContext(ctx, "DELETE", absUrl, nil)

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

func (c *Client) resourceGet(ctx context.Context, path string, out any) error {
	relUrl, err := url.Parse(path)

	if err != nil {
		return err
	}

	absUrl := c.urlBase.ResolveReference(relUrl).String()
	req, err := http.NewRequestWithContext(ctx, "GET", absUrl, nil)

	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	err = readJson(resp, &out)

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) resourceSignal(ctx context.Context, path string) error {
	relUrl, err := url.Parse(path)

	if err != nil {
		return err
	}

	absUrl := c.urlBase.ResolveReference(relUrl).String()
	req, err := http.NewRequestWithContext(ctx, "POST", absUrl, nil)

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

func (c *Client) resourceStream(ctx context.Context, path string, contentType string, reader io.Reader) error {
	relUrl, err := url.Parse(path)

	if err != nil {
		return err
	}

	absUrl := c.urlBase.ResolveReference(relUrl).String()
	req, err := http.NewRequestWithContext(ctx, "PUT", absUrl, reader)

	if err != nil {
		return err
	}

	req.Header.Add("content-type", contentType)
	resp, err := c.http.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return checkStatus(resp)
}

package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"sync"

	"github.com/decafcode/terraform-provider-podman/internal/client"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type PodmanProviderEnv struct {
	ContainerHost string
	SshAuthSock   string
}

type podmanProviderState struct {
	DefaultHost       string
	HostKeyAlgorithms []string

	mutex    sync.Mutex
	hosts    map[string]*client.Client
	sshAgent agent.ExtendedAgent
}

type publicKeyMismatchError struct {
	expectedKey string
	actualKey   string
}

func (e publicKeyMismatchError) Error() string {
	return fmt.Sprintf(
		"public key mismatch!\nExpected : %s\nActual   : %s\n",
		e.expectedKey,
		e.actualKey)
}

func newProviderState(env *PodmanProviderEnv) (*podmanProviderState, error) {
	var sshAgent agent.ExtendedAgent

	if env.SshAuthSock != "" {
		agentConn, err := net.Dial("unix", env.SshAuthSock)

		if err != nil {
			return nil, err
		}

		sshAgent = agent.NewClient(agentConn)
	}

	return &podmanProviderState{
		hosts:    make(map[string]*client.Client),
		sshAgent: sshAgent,
	}, nil
}

func (d *podmanProviderState) getClient(ctx context.Context, host string) (*client.Client, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if host == "" {
		if d.DefaultHost == "" {
			return nil, fmt.Errorf(
				"resource must specify container_host if neither a default container_host is specified in the provider nor a CONTAINER_HOST environment variable is set")
		}

		host = d.DefaultHost
	}

	existing := d.hosts[host]

	if existing != nil {
		return existing, nil
	}

	u, err := url.Parse(host)

	if err != nil {
		return nil, err
	}

	var authMethods []ssh.AuthMethod

	if d.sshAgent != nil {
		authMethods = append(authMethods, ssh.PublicKeysCallback(d.sshAgent.Signers))
	}

	config := client.Config{
		Ssh: ssh.ClientConfig{
			Auth:              authMethods,
			HostKeyAlgorithms: d.HostKeyAlgorithms,
		},
	}

	if u.Scheme == "ssh" {
		values, err := url.ParseQuery(u.Fragment)

		if err != nil {
			return nil, err
		}

		expectedKey := values.Get("pubkey")

		if expectedKey == "" {
			tflog.Warn(
				ctx,
				"SSH host public key was not specified, this is strongly discouraged and highly insecure! Please add a #pubkey=... URL fragment parameter to this URL.",
			)

			config.Ssh.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		} else {
			config.Ssh.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				actualKey := key.Type() + " " + base64.StdEncoding.EncodeToString(key.Marshal())

				// This is not a timing-safe comparison, but public keys are not generally
				// considered to be a secret to begin with so this is probably fine.

				if expectedKey != actualKey {
					return publicKeyMismatchError{expectedKey, actualKey}
				}

				return nil
			}
		}
	}

	c, err := client.Connect(ctx, u, &config)

	if err != nil {
		return nil, err
	}

	err = c.Ping(ctx)

	if err != nil {
		return nil, err
	}

	d.hosts[host] = c

	return c, nil
}

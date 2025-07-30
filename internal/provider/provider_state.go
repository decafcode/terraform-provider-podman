package provider

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"sync"

	"github.com/decafcode/terraform-provider-podman/internal/client"
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
			HostKeyCallback:   ssh.InsecureIgnoreHostKey(),
		},
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

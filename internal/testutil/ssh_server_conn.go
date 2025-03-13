package testutil

import (
	"fmt"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

type sshUnixChannelPath struct {
	Path      string
	Reserved1 string
	Reserved2 uint32
}

type sshServerConn struct {
	Conn net.Conn

	closed     bool
	serverConn *ssh.ServerConn
	newChans   <-chan ssh.NewChannel
}

func (c *sshServerConn) Close() error {
	if c.closed {
		return nil
	}

	var err error

	if c.serverConn != nil {
		err = c.serverConn.Close()
	}

	c.closed = true

	return err
}

func (c *sshServerConn) IsClosed() bool {
	return c.closed
}

func (c *sshServerConn) Start(conn net.Conn, cfg *ssh.ServerConfig) error {
	serverConn, newChans, reqs, err := ssh.NewServerConn(conn, cfg)

	if err != nil {
		return err
	}

	go ssh.DiscardRequests(reqs)
	c.serverConn = serverConn
	c.newChans = newChans

	return nil
}

func (c *sshServerConn) Accept(expectedPath string) (*sshServerTunnel, error) {
	for newChan := range c.newChans {
		acceptedChan, err := c.tryAccept(newChan, expectedPath)

		if err != nil {
			log.Println(err)
		} else {
			return &sshServerTunnel{Channel: acceptedChan}, nil
		}
	}

	return nil, fmt.Errorf("tunnel acceptance has ended")
}

func (c *sshServerConn) tryAccept(newChan ssh.NewChannel, expectedPath string) (ssh.Channel, error) {
	channelType := newChan.ChannelType()

	if channelType != "direct-streamlocal@openssh.com" {
		err := newChan.Reject(ssh.UnknownChannelType, "unexpected channel type")

		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("unexpected channel type %s", channelType)
	}

	path := sshUnixChannelPath{}
	err := ssh.Unmarshal(newChan.ExtraData(), &path)

	if err != nil {
		e2 := newChan.Reject(ssh.ConnectionFailed, "deserialization failure")

		if e2 != nil {
			return nil, e2
		}

		return nil, err
	}

	if path.Path != expectedPath {
		err := newChan.Reject(ssh.ConnectionFailed, "path mismatch")

		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("expected path \"%s\" got \"%s\"", expectedPath, path.Path)
	}

	acceptedChan, chanRequests, err := newChan.Accept()

	if err != nil {
		return nil, err
	}

	go ssh.DiscardRequests(chanRequests)

	return acceptedChan, nil
}

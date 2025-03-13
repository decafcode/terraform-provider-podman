package testutil

import (
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type sshServerTunnel struct {
	ssh.Channel

	mutex  sync.Mutex
	closed bool
}

func (t *sshServerTunnel) Close() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.Channel.Close()
}

func (t *sshServerTunnel) IsClosed() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.closed
}

func (t *sshServerTunnel) LocalAddr() net.Addr {
	return nil
}

func (t *sshServerTunnel) RemoteAddr() net.Addr {
	return nil
}

func (t *sshServerTunnel) SetDeadline(time.Time) error {
	return nil
}

func (t *sshServerTunnel) SetReadDeadline(time.Time) error {
	return nil
}

func (t *sshServerTunnel) SetWriteDeadline(time.Time) error {
	return nil
}

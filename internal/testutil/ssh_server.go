package testutil

import (
	"bytes"
	"fmt"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

type sshServerState int

const (
	stateNew = iota
	stateServing
	stateClosed
)

type SshServer struct {
	SocketPath     string
	HostPrivateKey ssh.Signer
	PublicKey      ssh.PublicKey
	Password       *string

	mutex     sync.Mutex
	state     sshServerState
	listener  net.Listener
	conns     []*sshServerConn
	tunnels   []*sshServerTunnel
	newTuns   chan *sshServerTunnel
	waitGroup sync.WaitGroup
}

func (s *SshServer) Close() error {
	var result error

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.state != stateServing {
		return nil
	}

	err := s.listener.Close()

	if err != nil {
		result = err
	}

	for _, tun := range s.tunnels {
		err := tun.Close()

		if err != nil && result == nil {
			result = err
		}
	}

	for _, conn := range s.conns {
		var err error

		if conn != nil {
			err = conn.Close()
		}

		if err != nil && result == nil {
			result = err
		}
	}

	s.conns = nil
	s.tunnels = nil
	s.state = stateClosed

	return result
}

func (s *SshServer) Serve(listener net.Listener) error {
	cfg := s.makeConfig()

	newTuns := make(chan *sshServerTunnel)
	defer close(newTuns)

	err := s.startServing(listener, newTuns)

	if err != nil {
		return err
	}

	defer s.waitGroup.Wait()

	for {
		tcpConn, err := listener.Accept()

		if err != nil {
			return err
		}

		s.waitGroup.Add(1)
		go s.serviceClientTask(tcpConn, cfg)
	}
}

func (s *SshServer) makeConfig() *ssh.ServerConfig {
	cfg := &ssh.ServerConfig{}
	cfg.AddHostKey(s.HostPrivateKey)

	if s.Password != nil {
		expected := []byte(*s.Password)

		cfg.PasswordCallback = func(_ ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			// Not timing safe but this is a test scaffold so we don't care
			if !bytes.Equal(expected, password) {
				return nil, fmt.Errorf("passwords do not match")
			}

			return nil, nil
		}
	}

	if s.PublicKey != nil {
		cfg.PublicKeyCallback = func(_ ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if !bytes.Equal(s.PublicKey.Marshal(), key.Marshal()) {
				return nil, fmt.Errorf("public keys do not match")
			}

			return nil, nil
		}
	}

	return cfg
}

func (s *SshServer) startServing(listener net.Listener, newTuns chan *sshServerTunnel) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.state != stateNew {
		return fmt.Errorf("unexpected state %d", s.state)
	}

	s.listener = listener
	s.newTuns = newTuns
	s.state = stateServing

	return nil
}

func (s *SshServer) serviceClientTask(tcpConn net.Conn, cfg *ssh.ServerConfig) {
	err := s.serviceClient(tcpConn, cfg)

	if err != nil {
		fmt.Println(err)
	}

	s.waitGroup.Done()
}

func (s *SshServer) serviceClient(tcpConn net.Conn, cfg *ssh.ServerConfig) error {
	sshConn := &sshServerConn{Conn: tcpConn}
	err := s.appendConn(sshConn)

	if err != nil {
		return err
	}

	defer s.removeConn(sshConn)
	err = sshConn.Start(tcpConn, cfg)

	if err != nil {
		return err
	}

	for {
		t, err := sshConn.Accept(s.SocketPath)

		if err != nil {
			return err
		}

		s.newTuns <- t
	}
}

func (s *SshServer) Accept() (net.Conn, error) {
	err := s.ensureStateServing()

	if err != nil {
		return nil, err
	}

	t := <-s.newTuns

	if t == nil {
		return nil, fmt.Errorf("server is closed")
	}

	err = s.appendTunnel(t)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *SshServer) Addr() net.Addr {
	return nil
}

func (s *SshServer) ensureStateServing() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.state != stateServing {
		return fmt.Errorf("server is not serving")
	}

	return nil
}

func (s *SshServer) appendConn(sshConn *sshServerConn) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.state != stateServing {
		return sshConn.Close()
	}

	for i := range s.conns {
		if s.conns[i] == nil {
			s.conns[i] = sshConn

			return nil
		}
	}

	s.conns = append(s.conns, sshConn)

	return nil
}

func (s *SshServer) removeConn(c *sshServerConn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := range s.conns {
		if s.conns[i] == c {
			s.conns[i] = nil
		}
	}
}

func (s *SshServer) appendTunnel(t *sshServerTunnel) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.state != stateServing {
		return t.Close()
	}

	for i := range s.tunnels {
		if s.tunnels[i].IsClosed() {
			s.tunnels[i] = t

			return nil
		}
	}

	s.tunnels = append(s.tunnels, t)

	return nil
}

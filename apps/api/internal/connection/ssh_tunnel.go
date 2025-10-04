package connection

import (
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHTunnel struct {
	Local  *net.TCPAddr
	Server *net.TCPAddr
	Remote *net.TCPAddr
	Config *ssh.ClientConfig
	client *ssh.Client
}

// NewSSHTunnel creates a new SSH tunnel configuration
func NewSSHTunnel(sshHost string, sshPort int, sshUsername, sshPassword, sshPrivateKey string, remoteHost string, remotePort int) (*SSHTunnel, error) {
	serverAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", sshHost, sshPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve SSH server address: %w", err)
	}

	// Remote database address (from SSH server's perspective)
	remoteAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", remoteHost, remotePort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve remote address: %w", err)
	}

	// Local address (random port on localhost)
	localAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve local address: %w", err)
	}

	var authMethods []ssh.AuthMethod

	if sshPassword != "" {
		authMethods = append(authMethods, ssh.Password(sshPassword))
	}

	if sshPrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(sshPrivateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no SSH authentication method provided (password or private key required)")
	}

	config := &ssh.ClientConfig{
		User:            sshUsername,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Add proper host key verification
		Timeout:         10 * time.Second,
	}

	return &SSHTunnel{
		Local:  localAddr,
		Server: serverAddr,
		Remote: remoteAddr,
		Config: config,
	}, nil
}

// Start establishes the SSH tunnel
func (tunnel *SSHTunnel) Start() error {
	client, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		return fmt.Errorf("failed to dial SSH server: %w", err)
	}
	tunnel.client = client

	// Listen on local port
	listener, err := net.ListenTCP("tcp", tunnel.Local)
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to listen on local port: %w", err)
	}

	// Update local address with actual port
	tunnel.Local = listener.Addr().(*net.TCPAddr)

	// Start forwarding in background
	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go tunnel.forward(conn)
		}
	}()

	return nil
}

// forward handles forwarding of a single connection
func (tunnel *SSHTunnel) forward(localConn net.Conn) {
	remoteConn, err := tunnel.client.Dial("tcp", tunnel.Remote.String())
	if err != nil {
		localConn.Close()
		return
	}

	// Copy data bidirectionally
	go func() {
		defer localConn.Close()
		defer remoteConn.Close()
		io.Copy(remoteConn, localConn)
	}()

	go func() {
		defer localConn.Close()
		defer remoteConn.Close()
		io.Copy(localConn, remoteConn)
	}()
}

// Stop closes the SSH tunnel
func (tunnel *SSHTunnel) Stop() error {
	if tunnel.client != nil {
		return tunnel.client.Close()
	}
	return nil
}

// GetLocalAddr returns the local address (host:port) to connect to
func (tunnel *SSHTunnel) GetLocalAddr() string {
	return tunnel.Local.String()
}

// GetLocalPort returns just the local port number
func (tunnel *SSHTunnel) GetLocalPort() int {
	return tunnel.Local.Port
}

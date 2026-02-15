package ssh

import (
	"fmt"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// AgentSigners returns signers from the running SSH agent via SSH_AUTH_SOCK.
// Returns nil, nil if SSH_AUTH_SOCK is not set (agent unavailable).
func AgentSigners() ([]ssh.Signer, error) {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return nil, nil
	}
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil, fmt.Errorf("connect to ssh-agent: %w", err)
	}
	defer conn.Close()

	ag := agent.NewClient(conn)
	signers, err := ag.Signers()
	if err != nil {
		return nil, fmt.Errorf("get agent signers: %w", err)
	}
	return signers, nil
}

// agentClient returns a connected agent client, or nil if unavailable.
func agentClient() (agent.ExtendedAgent, net.Conn, error) {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return nil, nil, nil
	}
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to ssh-agent: %w", err)
	}
	return agent.NewClient(conn), conn, nil
}

// setupAgentForwarding enables agent forwarding on an SSH session.
// The caller must close agentConn when done.
func setupAgentForwarding(client *ssh.Client) (agentConn net.Conn, err error) {
	ag, conn, err := agentClient()
	if err != nil {
		return nil, err
	}
	if ag == nil {
		return nil, nil
	}
	if err := agent.ForwardToAgent(client, ag); err != nil {
		conn.Close()
		return nil, fmt.Errorf("forward agent: %w", err)
	}
	return conn, nil
}

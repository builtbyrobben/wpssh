package ssh

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// LoadKey reads a private key file and parses it into an ssh.Signer.
// Supports both PEM (RSA, ECDSA, Ed25519) and OpenSSH private key formats.
func LoadKey(path string) (ssh.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key %s: %w", path, err)
	}
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("parse key %s: %w", path, err)
	}
	return signer, nil
}

// LoadKeyWithPassphrase reads an encrypted private key file and decrypts it.
func LoadKeyWithPassphrase(path, passphrase string) (ssh.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key %s: %w", path, err)
	}
	signer, err := ssh.ParsePrivateKeyWithPassphrase(data, []byte(passphrase))
	if err != nil {
		return nil, fmt.Errorf("parse encrypted key %s: %w", path, err)
	}
	return signer, nil
}

// Generate a new SSH key pair.
package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

// Generate a new SSH key pair.
type SshKeygen struct{}

type KeyPair struct {
	PublicKey  *File
	PrivateKey *Secret
}

// Generate a new SSH key pair using the Ed25519 algorithm.
func (m *SshKeygen) Ed25519() *Ed25519 {
	return &Ed25519{}
}

type Ed25519 struct{}

// Generate a new SSH key pair using the Ed25519 algorithm.
func (m *Ed25519) Generate(
	ctx context.Context,

	// Name used as public key file name and private key secret name.
	//
	// Set this to something unique if you need multiple keys.
	//
	// +optional
	// +default="id_ed25519"
	name string,

	// Encrypt the private key with the given passphrase.
	//
	// +optional
	passphrase *Secret,
) (*KeyPair, error) {
	cryptoPublicKey, cryptoPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	sshPublicKey, err := ssh.NewPublicKey(cryptoPublicKey)
	if err != nil {
		return nil, err
	}

	var sshPrivateKey *pem.Block
	{
		var err error

		if passphrase != nil {
			pass, err := passphrase.Plaintext(ctx)

			sshPrivateKey, err = ssh.MarshalPrivateKeyWithPassphrase(cryptoPrivateKey, "", []byte(pass))
			if err != nil {
				return nil, err
			}
		} else {
			sshPrivateKey, err = ssh.MarshalPrivateKey(cryptoPrivateKey, "")
			if err != nil {
				return nil, err
			}
		}
	}

	publicKey := dag.Directory().WithNewFile(name+".pub", string(ssh.MarshalAuthorizedKey(sshPublicKey))).File(name + ".pub")
	privateKey := dag.SetSecret(name, string(pem.EncodeToMemory(sshPrivateKey)))

	return &KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

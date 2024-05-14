// Generate a new SSH key pair.
package main

import (
	"context"
	"crypto"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

// Generate a new SSH key pair.
type SshKeygen struct{}

// An SSH key pair.
type KeyPair struct {
	PublicKey  *File
	PrivateKey *Secret
}

func formatKeys(
	ctx context.Context,
	name string,
	publicKey crypto.PublicKey,
	privateKey crypto.PrivateKey,
	passphrase *Secret,
) (*KeyPair, error) {
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	var sshPrivateKey *pem.Block
	{
		var err error

		if passphrase != nil {
			pass, err := passphrase.Plaintext(ctx)
			if err != nil {
				return nil, err
			}

			sshPrivateKey, err = ssh.MarshalPrivateKeyWithPassphrase(privateKey, "", []byte(pass))
			if err != nil {
				return nil, err
			}
		} else {
			sshPrivateKey, err = ssh.MarshalPrivateKey(privateKey, "")
			if err != nil {
				return nil, err
			}
		}
	}

	return &KeyPair{
		PublicKey:  dag.Directory().WithNewFile(name+".pub", string(ssh.MarshalAuthorizedKey(sshPublicKey))).File(name + ".pub"),
		PrivateKey: dag.SetSecret(name, string(pem.EncodeToMemory(sshPrivateKey))),
	}, nil
}

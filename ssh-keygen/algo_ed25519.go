package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
)

// Generate a new SSH key pair using the Ed25519 algorithm.
func (m *SshKeygen) Ed25519() *Ed25519 {
	return &Ed25519{}
}

// Generate a new SSH key pair using the Ed25519 algorithm.
type Ed25519 struct{}

// Generate a new SSH key pair using the Ed25519 algorithm.
func (*Ed25519) Generate(
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
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return formatKeys(ctx, name, publicKey, privateKey, passphrase)
}

package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
)

// Generate a new SSH key pair using the RSA algorithm.
func (m *SshKeygen) Rsa(
	// Generate an RSA private key with the given number of bits.
	//
	// +optional
	// +default=4096
	bits int,
) *Rsa {
	return &Rsa{
		Bits: bits,
	}
}

// Generate a new SSH key pair using the RSA algorithm.
type Rsa struct {
	Bits int
}

// Generate a new SSH key pair using the RSA algorithm.
func (algo *Rsa) Generate(
	ctx context.Context,

	// Name used as public key file name and private key secret name.
	//
	// Set this to something unique if you need multiple keys.
	//
	// +optional
	// +default="id_rsa"
	name string,

	// Encrypt the private key with the given passphrase.
	//
	// +optional
	passphrase *Secret,
) (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, algo.Bits)
	if err != nil {
		return nil, err
	}

	return formatKeys(ctx, name, privateKey.Public(), privateKey, passphrase)
}

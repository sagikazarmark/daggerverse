package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"dagger/ssh-keygen/internal/dagger"
	"fmt"
	"slices"
)

// Generate a new SSH key pair using the ECDSA algorithm.
func (m *SshKeygen) Ecdsa(
	// Generate an ECDSA private key with the given number of bits.
	//
	// +optional
	// +default=256
	bits int,
) (*Ecdsa, error) {
	if !slices.Contains([]int{256, 384, 521}, bits) {
		return nil, fmt.Errorf("invalid number of bits: %d", bits)
	}

	return &Ecdsa{
		Bits: bits,
	}, nil
}

// Generate a new SSH key pair using the ECDSA algorithm.
type Ecdsa struct {
	Bits int
}

// Generate a new SSH key pair using the ECDSA algorithm.
func (algo *Ecdsa) Generate(
	ctx context.Context,

	// Name used as public key file name and private key secret name.
	//
	// Set this to something unique if you need multiple keys.
	//
	// +optional
	// +default="id_ecdsa"
	name string,

	// Encrypt the private key with the given passphrase.
	//
	// +optional
	passphrase *dagger.Secret,
) (*KeyPair, error) {
	var curve elliptic.Curve

	switch algo.Bits {
	case 256:
		curve = elliptic.P256()

	case 384:
		curve = elliptic.P384()

	case 521:
		curve = elliptic.P521()

	default:
		panic("invalid number of bits")
	}

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}

	return formatKeys(ctx, name, privateKey.Public(), privateKey, passphrase)
}

package main

import (
	"context"
	"dagger/ssh-keygen/tests/internal/dagger"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Ed25519)
	p.Go(m.Rsa)
	p.Go(m.Ecdsa)

	return p.Wait()
}

func (m *Tests) Ed25519(ctx context.Context) error {
	keyPair := dag.SSHKeygen().Ed25519().Generate()

	return verify(ctx, keyPair)
}

func (m *Tests) Rsa(ctx context.Context) error {
	keyPair := dag.SSHKeygen().Rsa().Generate()

	return verify(ctx, keyPair)
}

func (m *Tests) Ecdsa(ctx context.Context) error {
	keyPair := dag.SSHKeygen().Ecdsa().Generate()

	return verify(ctx, keyPair)
}

func verify(ctx context.Context, keyPair *dagger.SSHKeygenKeyPair) error {
	_, err := dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		WithExec([]string{"apk", "add", "openssh-keygen"}).
		WithMountedFile("/ssh/public", keyPair.PublicKey()).
		WithMountedSecret("/ssh/private", keyPair.PrivateKey()).
		WithExec([]string{"ssh-keygen", "-l", "-f", "/ssh/private"}).
		WithExec([]string{"ssh-keygen", "-l", "-f", "/ssh/public"}).
		WithExec([]string{"sh", "-c", "diff -s <(ssh-keygen -l -f /ssh/private | cut -d' ' -f2) <(ssh-keygen -l -f /ssh/public | cut -d' ' -f2)"}).
		Sync(ctx)

	return err
}

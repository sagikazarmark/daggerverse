package main

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Ed25519)

	return p.Wait()
}

func (m *Tests) Ed25519(ctx context.Context) error {
	keyPair := dag.SSHKeygen().Ed25519().Generate()

	_, err := keyPair.PublicKey().Sync(ctx)

	return err
}

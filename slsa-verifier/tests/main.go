package main

import (
	"context"
	"fmt"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Verify)

	return p.Wait()
}

func (m *Tests) Verify(ctx context.Context) error {
	const version = "2.4.1"

	artifact := dag.HTTP(fmt.Sprintf("https://github.com/slsa-framework/slsa-verifier/releases/download/v%s/slsa-verifier-linux-amd64", version))

	provenance := dag.HTTP(fmt.Sprintf("https://github.com/slsa-framework/slsa-verifier/releases/download/v%s/slsa-verifier-linux-amd64.intoto.jsonl", version))

	_, err := dag.SlsaVerifier().VerifyArtifact(
		[]*File{artifact},
		provenance,
		"github.com/slsa-framework/slsa-verifier",
		SlsaVerifierVerifyArtifactOpts{
			SourceTag: "v2.4.1",
		},
	).Sync(ctx)

	return err
}

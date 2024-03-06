package main

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

func (m *Ci) SlsaVerifier(ctx context.Context) error {
	var group errgroup.Group

	// Verify artifact
	group.Go(func() error {
		const slsaVerifierVersion = "2.4.1"

		artifact := dag.Container().
			From("alpine:latest").
			WithExec([]string{
				"wget",
				"-O", "/tmp/slsa-verifier-linux-amd64",
				fmt.Sprintf("https://github.com/slsa-framework/slsa-verifier/releases/download/v%s/slsa-verifier-linux-amd64", slsaVerifierVersion),
			}).File("/tmp/slsa-verifier-linux-amd64")

		provenance := dag.Container().
			From("alpine:latest").
			WithExec([]string{
				"wget",
				"-O", "/tmp/slsa-verifier-linux-amd64.intoto.jsonl",
				fmt.Sprintf("https://github.com/slsa-framework/slsa-verifier/releases/download/v%s/slsa-verifier-linux-amd64.intoto.jsonl", slsaVerifierVersion),
			}).File("/tmp/slsa-verifier-linux-amd64.intoto.jsonl")

		_, err := dag.SlsaVerifier().VerifyArtifact(
			[]*File{artifact},
			provenance,
			"github.com/slsa-framework/slsa-verifier",
			SlsaVerifierVerifyArtifactOpts{
				SourceTag: "v2.4.1",
			},
		).Sync(ctx)

		return err
	})

	return group.Wait()
}

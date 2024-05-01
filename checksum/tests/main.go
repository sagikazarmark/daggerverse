package main

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.CalculateAndCheck)

	return p.Wait()
}

func (m *Tests) CalculateAndCheck(ctx context.Context) error {
	files := []*File{
		dag.CurrentModule().Source().File("./testdata/foo"),
		dag.CurrentModule().Source().File("./testdata/bar"),
	}

	checksums := dag.Checksum().Sha256().Calculate(files)

	_, err := dag.Checksum().Sha256().Check(checksums, files).Sync(ctx)

	return err
}

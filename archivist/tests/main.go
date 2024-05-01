package main

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.ArchiveDirectory)

	return p.Wait()
}

func (m *Tests) ArchiveDirectory(_ context.Context) error {
	dir := dag.CurrentModule().Source().Directory("./testdata")

	archive := dag.Archivist().TarGz().Archive("test", dir)

	unarchivedDir := dag.Arc().Unarchive(archive)

	// TODO: improve test
	_ = unarchivedDir

	return nil
}

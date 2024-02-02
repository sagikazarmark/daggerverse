package main

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func (m *Ci) Archivist(ctx context.Context) error {
	var group errgroup.Group

	group.Go(func() error {
		dir := dag.CurrentModule().Source().Directory("./testdata/archivist")

		archive := dag.Archivist().TarGz().Archive("test", dir)

		unarchivedDir := dag.Arc().Unarchive(archive)

		// TODO: improve test
		_ = unarchivedDir

		return nil
	})

	return group.Wait()
}

package main

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func (m *Ci) Arc(ctx context.Context) error {
	var group errgroup.Group

	group.Go(func() error {
		dir := dag.Host().Directory("./testdata/arc")

		archive := dag.Arc().ArchiveDirectory("test", dir).TarGz()

		unarchivedDir := dag.Arc().Unarchive(archive)

		// TODO: improve test

		_ = unarchivedDir

		return nil
	})

	return group.Wait()
}

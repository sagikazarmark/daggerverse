package main

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func (m *Ci) Archiver(ctx context.Context) error {
	var group errgroup.Group

	group.Go(func() error {
		dir := dag.Host().Directory("./testdata/archiver")

		archive := dag.Archiver().TarGz().Archive("test", dir)

		unarchivedDir := dag.Arc().Unarchive(archive)

		// TODO: improve test

		_ = unarchivedDir

		return nil
	})

	return group.Wait()
}

package main

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func (m *Ci) Checksum(ctx context.Context) error {
	var group errgroup.Group

	// Calculate and check files
	group.Go(func() error {
		files := []*File{
			dag.CurrentModule().Source().File("./testdata/checksum/foo"),
			dag.CurrentModule().Source().File("./testdata/checksum/bar"),
		}

		checksums := dag.Checksum().Sha256().Calculate(files)

		_, err := dag.Checksum().Sha256().Check(checksums, files).Sync(ctx)

		return err
	})

	return group.Wait()
}

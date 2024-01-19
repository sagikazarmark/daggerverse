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
			dag.Host().File("./testdata/checksum/foo"),
			dag.Host().File("./testdata/checksum/bar"),
		}

		checksums, err := dag.Checksum().Sha256(ctx, files)
		if err != nil {
			return err
		}

		_, err = dag.Checksum().CheckSha256(checksums, files).Sync(ctx)

		return err
	})

	return group.Wait()
}

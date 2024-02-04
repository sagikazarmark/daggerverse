package main

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func (m *Ci) Quarto(ctx context.Context) error {
	var group errgroup.Group

	group.Go(func() error {
		dir := dag.CurrentModule().Source().Directory("./testdata/quarto")

		_, err := dag.Quarto().Render(dir).Directory().Sync(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	return group.Wait()
}

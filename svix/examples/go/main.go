// Examples for Svix module.

package main

import (
	"context"
	"dagger/svix/examples/go/internal/dagger"

	"github.com/sourcegraph/conc/pool"
)

type Examples struct{}

// All executes all examples.
func (m *Examples) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Svix_Defaults)
	p.Go(m.Svix_Postgres)

	return p.Wait()
}

func (m *Examples) Svix_Defaults(ctx context.Context) error {
	svix := dag.Svix()

	_, err := svix.Service().Start(ctx)

	return err
}

func (m *Examples) Svix_Postgres(ctx context.Context) error {
	postgres := dag.Postgres(dagger.PostgresOpts{
		User:     dag.SetSecret("postgres-user", "postgres"),
		Password: dag.SetSecret("postgres-password", "postgres"),
	})

	svix := dag.Svix(dagger.SvixOpts{
		Postgres: postgres.AsSvixPostgres(),
	})

	_, err := svix.Service().Start(ctx)

	return err
}

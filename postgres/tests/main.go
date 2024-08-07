package main

import (
	"context"
	"dagger/postgres/tests/internal/dagger"

	"github.com/sourcegraph/conc/pool"
)

const postgresVersion = "16.3"

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Basic)

	return p.Wait()
}

func (m *Tests) Basic(ctx context.Context) error {
	postgres := dag.Postgres(dagger.PostgresOpts{
		Version: postgresVersion,
	})

	client, err := client(ctx, postgres)
	if err != nil {
		return err
	}

	_, err = client.RunCommand(ctx, "SELECT 1;")
	if err != nil {
		return err
	}

	return nil
}

func client(ctx context.Context, postgres *dagger.Postgres) (*dagger.Psql, error) {
	user := postgres.User()
	password := postgres.Password()

	database, err := postgres.Database(ctx)
	if err != nil {
		return nil, err
	}

	return dag.Psql(dagger.PsqlOpts{
		Version:  postgresVersion,
		Service:  postgres.Service(),
		User:     user,
		Password: password,
		Database: database,
	}), nil
}

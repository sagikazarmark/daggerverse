package main

import (
	"context"
	"dagger/postgres/tests/internal/dagger"
	"errors"
	"slices"

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

func (m *Tests) WithDatabase(ctx context.Context) error {
	postgres := dag.Postgres(dagger.PostgresOpts{
		Version: postgresVersion,
	}).WithDatabase("test")

	client, err := client(ctx, postgres)
	if err != nil {
		return err
	}

	databases, err := client.List(ctx)
	if err != nil {
		return err
	}

	databaseNames := make([]string, 0, len(databases))

	for _, database := range databases {
		name, err := database.Name(ctx)
		if err != nil {
			return err
		}

		databaseNames = append(databaseNames, name)
	}

	if !slices.Contains(databaseNames, "test") {
		return errors.New("expected to find test database")
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

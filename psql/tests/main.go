package main

import (
	"context"
	"dagger/psql/tests/internal/dagger"
	"fmt"

	"github.com/sourcegraph/conc/pool"
)

const postgresVersion = "16.3"

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.List)
	p.Go(m.RunCommand)
	p.Go(m.RunFile)
	p.Go(m.Run)

	return p.Wait()
}

func newPsql() *dagger.Psql {
	user := dag.SetSecret("postgres-user", "foo")
	password := dag.SetSecret("postgres-password", "bar")
	database := "postgres"

	postgres := dag.Postgres(dagger.PostgresOpts{
		Version:  postgresVersion,
		User:     user,
		Password: password,
		Database: database,
	})

	return dag.Psql(dagger.PsqlOpts{
		Version:  postgresVersion,
		Service:  postgres.Service(),
		User:     user,
		Password: password,
	})
}

func (m *Tests) List(ctx context.Context) error {
	psql := newPsql()

	list, err := psql.List(ctx)
	if err != nil {
		return err
	}

	if len(list) < 3 {
		return fmt.Errorf("expected at least 3 databases, got %d", len(list))
	}

	return nil
}

func (m *Tests) RunCommand(ctx context.Context) error {
	psql := newPsql()

	_, err := psql.RunCommand(ctx, "SELECT 1;")
	if err != nil {
		return err
	}

	return nil
}

func (m *Tests) RunFile(ctx context.Context) error {
	psql := newPsql()

	_, err := psql.RunFile(ctx, dag.Directory().WithNewFile("command", "SELECT 1;").File("command"))
	if err != nil {
		return err
	}

	return nil
}

func (m *Tests) Run(ctx context.Context) error {
	psql := newPsql()

	run := psql.Run().
		WithCommand("SELECT 1;").
		WithFile(dag.Directory().WithNewFile("command", "SELECT 1;").File("command"))

	_, err := run.Execute(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Launch a Postgres database server.
//
// This module is aimed at providing a simple way to launch a Postgres database server for development and testing purposes.
//
// Check out the official Postgres image on Docker Hub for further customization options: https://hub.docker.com/_/postgres

package main

import (
	"context"
	"dagger/postgres/internal/dagger"
	"fmt"
	"strings"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "postgres"

type Postgres struct {
	// Superuser name.
	User *dagger.Secret

	// Superuser password.
	Password *dagger.Secret

	// The database name.
	Database string

	// +private
	Container *dagger.Container

	// +private
	InitScripts *dagger.Directory
}

func New(
	ctx context.Context,

	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	version string,

	// Custom container to use as a base container. Takes precedence over version.
	//
	// +optional
	container *dagger.Container,

	// Superuser name. (default "postgres")
	//
	// +optional
	user *dagger.Secret,

	// Superuser password. (defaults "postgres")
	//
	// +optional
	password *dagger.Secret,

	// The database name. (defaults to the user name)
	//
	// +optional
	database string,

	// Mount a volume to persist data between runs.
	//
	// +optional
	dataVolume *dagger.CacheVolume,

	// Initialization scripts (*.sql and *.sh) to run when the service first starts.
	//
	// +optional
	initScripts *dagger.Directory,

	// Additional arguments to pass to initdb.
	//
	// +optional
	initdbArgs []string,

	// Instance name (allowing to spawn multiple services with the same parameters).
	//
	// +optional
	name string,
) (*Postgres, error) {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(defaultImageRepository + ":" + version)
	}

	// User defaults to "postgres".
	if user == nil {
		user = dag.SetSecret("postgres-default-user", "postgres")
	}

	// Password defaults to "postgres".
	if password == nil {
		password = dag.SetSecret("postgres-default-password", "postgres")
	}

	// Database defaults to the user name.
	if database == "" {
		var err error

		database, err = user.Plaintext(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Support custom PGDATA coming from a custom container.
	var pgdata string

	if dataVolume != nil {
		var err error

		pgdata, err = container.EnvVariable(ctx, "PGDATA")
		if err != nil {
			return nil, err
		}

		if pgdata == "" {
			pgdata = "/var/lib/postgresql/data"
		}
	}

	if initScripts == nil {
		initScripts = dag.Directory()
	}

	container = container.
		WithExposedPort(5432).
		WithSecretVariable("POSTGRES_USER", user).
		WithSecretVariable("POSTGRES_PASSWORD", password).
		WithEnvVariable("POSTGRES_DB", database).
		With(func(c *dagger.Container) *dagger.Container {
			if dataVolume != nil {
				c = c.WithMountedCache(pgdata, dataVolume)
			}

			if len(initdbArgs) > 0 {
				c = c.WithEnvVariable("POSTGRES_INITDB_ARGS", strings.Join(initdbArgs, " "))
			}

			if name != "" {
				c = c.WithLabel("io.dagger.postgres.instance", name)
			}

			return c
		})

	return &Postgres{
		User:     user,
		Password: password,
		Database: database,

		Container:   container,
		InitScripts: initScripts,
	}, nil
}

func (m *Postgres) container() *dagger.Container {
	return m.Container.
		WithMountedDirectory("/docker-entrypoint-initdb.d", m.InitScripts)
}

// Add an additional initialization script to run when the service first starts.
func (m *Postgres) WithInitScript(file *dagger.File) *Postgres {
	m.InitScripts = m.InitScripts.WithFile("", file)

	return m
}

// Creates an additional database with the given name (with the default user as the owner) when the service first starts.
//
// Under the hood, this method adds a SQL script to the init scripts that creates the database.
func (m *Postgres) WithDatabase(
	ctx context.Context,

	// Database to create.
	name string,
) (*Postgres, error) {
	user, err := m.User.Plaintext(ctx)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`CREATE DATABASE %s; GRANT ALL PRIVILEGES ON DATABASE %s TO %s;`, name, name, user)

	m.InitScripts = m.InitScripts.WithNewFile(fmt.Sprintf("create-database-%s.sql", name), sql)

	return m, nil
}

// The Postgres service.
func (m *Postgres) Service() *dagger.Service {
	return m.container().AsService(dagger.ContainerAsServiceOpts{
		UseEntrypoint: true,
	})
}

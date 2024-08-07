// Launch a Postgres database server.
//
// This module is aimed at providing a simple way to launch a Postgres database server for development and testing purposes.
//
// Check out the official Postgres image on Docker Hub for further customization options: https://hub.docker.com/_/postgres

package main

import (
	"context"
	"crypto/sha1"
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

	// Superuser password. (defaults to a generated password)
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

	// Generate a random password.
	if password == nil {
		randomPassword, err := generateRandomPassword(20)
		if err != nil {
			return nil, err
		}

		h := sha1.New()

		_, err = h.Write([]byte(randomPassword))
		if err != nil {
			return nil, err
		}

		name := fmt.Sprintf("postgres-generated-password-%x", h.Sum(nil))

		password = dag.SetSecret(name, randomPassword)
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

	container = container.
		WithExposedPort(5432).
		WithSecretVariable("POSTGRES_USER", user).
		WithSecretVariable("POSTGRES_PASSWORD", password).
		WithEnvVariable("POSTGRES_DB", database).
		With(func(c *dagger.Container) *dagger.Container {
			if dataVolume != nil {
				c = c.WithMountedCache(pgdata, dataVolume)
			}

			if initScripts != nil {
				c = c.WithMountedDirectory("/docker-entrypoint-initdb.d", initScripts)
			}

			if len(initdbArgs) > 0 {
				c = c.WithEnvVariable("POSTGRES_INITDB_ARGS", strings.Join(initdbArgs, " "))
			}

			return c
		})

	return &Postgres{
		User:     user,
		Password: password,
		Database: database,

		Container: container,
	}, nil
}

// The Postgres service.
func (m *Postgres) Service() *dagger.Service {
	return m.Container.AsService()
}

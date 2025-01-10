// Svix is the enterprise ready webhook service.
//
// This module allows running Svix for development and testing purposes.

package main

import (
	"context"
	"dagger/svix/internal/dagger"
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "svix/svix-server"

type Svix struct {
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

	// Postgres service.
	//
	// +optional
	postgres Postgres,

	// Override the database name provided by the Postgres service.
	//
	// +optional
	database string,

	// The JWT secret for authentication. (defaults to a generated secret)
	//
	// +optional
	jwtSecret *dagger.Secret,

	// Svix configuration file.
	//
	// +optional
	config *dagger.File,
) (*Svix, error) {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(defaultImageRepository + ":" + version)
	}

	container = container.
		WithEnvVariable("SVIX_QUEUE_TYPE", "memory").
		WithEnvVariable("SVIX_CACHE_TYPE", "memory").
		WithEnvVariable("SVIX_LISTEN_ADDRESS", "0.0.0.0:8071").
		WithExposedPort(8071).
		WithEnvVariable("WAIT_FOR", "true")

	{
		if postgres == nil {
			postgres = dag.Postgres()
		}

		user, err := postgres.User().Plaintext(ctx)
		if err != nil {
			return nil, err
		}

		password, err := postgres.Password().Plaintext(ctx)
		if err != nil {
			return nil, err
		}

		if database == "" {
			var err error

			database, err = postgres.Database(ctx)
			if err != nil {
				return nil, err
			}
		}

		dsn := fmt.Sprintf("postgresql://%s:%s@postgres/%s?sslmode=disable", user, password, database)

		container = container.
			WithEnvVariable("SVIX_DB_DSN", dsn).
			WithServiceBinding("postgres", postgres.Service())
	}

	if jwtSecret == nil {
		var err error

		jwtSecret, err = generateRandomSecret("svix-generated-jwt-secret", 20)
		if err != nil {
			return nil, err
		}
	}

	container = container.
		WithSecretVariable("SVIX_JWT_SECRET", jwtSecret)

	if config != nil {
		container = container.
			WithMountedFile("/config.toml", config)
	}

	return &Svix{
		Container: container,
	}, nil
}

// Postgres service dependency.
//
// See the following module for a compatible implementation: https://daggerverse.dev/mod/github.com/sagikazarmark/daggerverse/postgres
type Postgres interface {
	dagger.DaggerObject

	User() *dagger.Secret
	Password() *dagger.Secret
	Database(ctx context.Context) (string, error)
	Service() *dagger.Service
}

func (m *Svix) Service() *dagger.Service {
	return m.Container.AsService(dagger.ContainerAsServiceOpts{
		UseEntrypoint: true,
	})
}

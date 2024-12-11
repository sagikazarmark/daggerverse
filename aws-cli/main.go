// AWS Command Line Interface
//
// This module offers a generic interface to use AWS CLI in Dagger
// as well as some high level functions that come in handy in the context of interacting with AWS from a Dagger pipeline.

package main

import (
	"dagger/aws-cli/internal/dagger"
	"fmt"
	"time"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "public.ecr.aws/aws-cli/aws-cli"

type AwsCli struct {
	Region string

	Container *dagger.Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	version string,

	// Custom container to use as a base container.
	//
	// +optional
	container *dagger.Container,

	// Default AWS region.
	//
	// +optional
	region string,
) AwsCli {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	if region != "" {
		container = container.WithEnvVariable("AWS_REGION", region)
	}

	m := AwsCli{
		Region:    region,
		Container: container,
	}

	return m
}

// Set a region for all AWS CLI commands.
func (m AwsCli) WithRegion(
	// AWS region.
	region string,
) AwsCli {
	m.Region = region
	m.Container = m.Container.
		WithEnvVariable("AWS_REGION", region)

	return m
}

// Mount an AWS CLI config file.
func (m AwsCli) WithConfig(
	// AWS config file.
	source *dagger.File,
) AwsCli {
	m.Container = m.Container.
		WithMountedFile("/root/.aws/config", source)

	return m
}

// Mount an AWS CLI credentials file.
func (m AwsCli) WithCredentials(
	// AWS credentials file.
	source *dagger.Secret,
) AwsCli {
	m.Container = m.Container.
		WithMountedSecret("/root/.aws/credentials", source)

	return m
}

// Set a profile for all AWS CLI commands.
//
// Should be used with (at least) WithConfig.
func (m AwsCli) WithProfile(
	// AWS profile.
	profile string,
) AwsCli {
	m.Container = m.Container.
		WithEnvVariable("AWS_PROFILE", profile)

	return m
}

// Set static AWS credentials.
func (m AwsCli) WithStaticCredentials(
	// AWS access key.
	accessKeyId *dagger.Secret,

	// AWS secret key.
	secretAccessKey *dagger.Secret,

	// AWS session token (for temporary credentials).
	//
	// +optional
	sessionToken *dagger.Secret,
) AwsCli {
	m.Container = m.Container.
		WithSecretVariable("AWS_ACCESS_KEY_ID", accessKeyId).
		WithSecretVariable("AWS_SECRET_ACCESS_KEY", secretAccessKey).
		With(func(c *dagger.Container) *dagger.Container {
			if sessionToken == nil {
				// make sure to reset any previously set value
				return c.
					WithoutEnvVariable("AWS_SESSION_TOKEN"). // make sure that no previous env vars are set in the container
					WithoutSecretVariable("AWS_SESSION_TOKEN")
			}

			return c.WithSecretVariable("AWS_SESSION_TOKEN", sessionToken)
		})

	return m
}

// Remove previously set static AWS credentials.
func (m AwsCli) WithoutStaticCredentials() AwsCli {
	m.Container = m.Container.
		// make sure that no previous env vars are set in the container
		WithoutEnvVariable("AWS_ACCESS_KEY_ID").
		WithoutEnvVariable("AWS_SECRET_ACCESS_KEY").
		WithoutEnvVariable("AWS_SESSION_TOKEN").
		WithoutSecretVariable("AWS_ACCESS_KEY_ID").
		WithoutSecretVariable("AWS_SECRET_ACCESS_KEY").
		WithoutSecretVariable("AWS_SESSION_TOKEN")

	return m
}

// Set static AWS credentials (shorthand for WithStaticCredentials, making the session token a required parameter).
func (m AwsCli) WithTemporaryCredentials(
	// AWS access key.
	accessKeyId *dagger.Secret,

	// AWS secret key.
	secretAccessKey *dagger.Secret,

	// AWS session token (for temporary credentials).
	sessionToken *dagger.Secret,
) AwsCli {
	return m.WithStaticCredentials(accessKeyId, secretAccessKey, sessionToken)
}

// Run an AWS CLI command.
func (m AwsCli) Exec(
	// Command to run (without "aws") (e.g., ["sts", "get-caller-identity"]).
	args []string,
) *dagger.Container {
	args = append([]string{"aws"}, args...)

	return m.Container.
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec(args)
}

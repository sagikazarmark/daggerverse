package main

import (
	"context"
	"crypto/sha1"
	"dagger/aws-cli/internal/dagger"
	"fmt"
	"time"
)

// Amazon Elastic Container Registry (Amazon ECR) is a managed container image registry service.
func (m AwsCli) Ecr() Ecr {
	return Ecr{
		Cli: m,
	}
}

type Ecr struct {
	// +private
	Cli AwsCli
}

// Run an AWS CLI command.
func (m Ecr) Exec(
	// Command to run (without "aws ecr").
	args []string,
) *dagger.Container {
	args = append([]string{"ecr"}, args...)

	return m.Cli.Exec(args)
}

// Retrieves an authentication token that you can use to authenticate to an Amazon ECR registry.
func (m Ecr) GetLoginPassword(
	ctx context.Context,

	// AWS region (required to be set either globally or here).
	//
	// +optional
	region string,

	// A Go compatible duration string to control how long the result of this call should stay in Dagger's cache.
	//
	// +optional
	// +default="12h"
	cacheTtl string,
) (*dagger.Secret, error) {
	args := []string{"get-login-password"}

	if region != "" {
		args = append(args, "--region", region)
	}

	if cacheTtl == "" { // just to be safe
		cacheTtl = "12h"
	}

	cacheTtlDuration, err := time.ParseDuration(cacheTtl)
	if err != nil {
		return nil, err
	}

	// this should be safe because receivers are not pointers
	m.Cli.Container = m.Cli.Container.WithEnvVariable("CACHE_BUSTER", time.Now().Truncate(cacheTtlDuration).Format(time.RFC3339Nano))

	password, err := m.Exec(args).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	h := sha1.New()

	_, err = h.Write([]byte(password))
	if err != nil {
		return nil, err
	}

	const prefix = "aws-ecr-get-login-password"

	name := fmt.Sprintf("%s-%x", prefix, h.Sum(nil))

	return dag.SetSecret(name, password), nil
}

// Create registry authentication details for an ECR registry.
func (m Ecr) Login(
	ctx context.Context,

	// Account ID to be used for assembling the registry address. (Falls back to the account ID of the caller)
	//
	// +optional
	accountId string,

	// AWS region (required to be set either globally or here).
	//
	// +optional
	region string,

	// A Go compatible duration string to control how long the password should stay in Dagger's cache.
	//
	// +optional
	// +default="12h"
	passwordCacheTtl string,
) (*EcrRegistry, error) {
	if accountId == "" {
		id, err := m.Cli.Sts().GetCallerIdentity(ctx)
		if err != nil {
			return nil, err
		}

		accountId = id.Account
	}

	password, err := m.GetLoginPassword(ctx, region, passwordCacheTtl)
	if err != nil {
		return nil, err
	}

	// use the global region (if any)
	// if there is no global region set, the password fetch fail anyway (hence no error returned directly)
	if region == "" {
		region = m.Cli.Region
	}

	address := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", accountId, region)

	return &EcrRegistry{
		Address:  address,
		Password: password,
	}, nil
}

type EcrRegistry struct {
	Address string

	Password *dagger.Secret
}

// Set registry authentication in a container.
func (m *EcrRegistry) Auth(container *dagger.Container) *dagger.Container {
	return container.WithRegistryAuth(m.Address, "AWS", m.Password)
}

package main

import (
	"context"
	"dagger/aws-cli/internal/dagger"
	"encoding/json"
)

// Amazon Elastic Container Registry (Amazon ECR) is a managed container image registry service.
func (m AwsCli) Sts() Sts {
	return Sts{
		Cli: m,
	}
}

type Sts struct {
	// +private
	Cli AwsCli
}

// Run an AWS CLI command.
func (m Sts) Exec(
	// Command to run (without "aws sts").
	args []string,
) *dagger.Container {
	args = append([]string{"sts"}, args...)

	return m.Cli.Exec(args)
}

// Returns  details  about the IAM user or role whose credentials are used to call the operation.
func (m Sts) GetCallerIdentity(ctx context.Context) (*StsCallerIdentity, error) {
	raw, err := m.Exec([]string{"get-caller-identity"}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	var id StsCallerIdentity

	err = json.Unmarshal([]byte(raw), &id)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

type StsCallerIdentity struct {
	UserId  string
	Account string
	Arn     string
}

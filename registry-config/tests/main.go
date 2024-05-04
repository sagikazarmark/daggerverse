package main

import (
	"context"
	"fmt"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.WithRegistryAuth)
	p.Go(m.WithRegistryAuth_MultipleCredentials)

	return p.Wait()
}

func (m *Tests) WithRegistryAuth(ctx context.Context) error {
	secret := dag.RegistryConfig().
		WithRegistryAuth("ghcr.io", "sagikazarmark", dag.SetSecret("password", "password")).
		WithRegistryAuth("docker.io", "sagikazarmark", dag.SetSecret("password2", "password2")).
		Secret()

	actual, err := secret.Plaintext(ctx)
	if err != nil {
		return err
	}

	const expected = `{"auths":{"docker.io":{"auth":"c2FnaWthemFybWFyazpwYXNzd29yZDI="},"ghcr.io":{"auth":"c2FnaWthemFybWFyazpwYXNzd29yZA=="}}}`

	if actual != expected {
		return fmt.Errorf("secret does not match the expected value\nactual:   %s\nexpected: %s", actual, expected)
	}

	return nil
}

func (m *Tests) WithRegistryAuth_MultipleCredentials(ctx context.Context) error {
	secret := dag.RegistryConfig().
		WithRegistryAuth("ghcr.io", "sagikazarmark", dag.SetSecret("passwordold", "passwordold")).
		WithRegistryAuth("docker.io", "sagikazarmark", dag.SetSecret("password2", "password2")).
		WithRegistryAuth("ghcr.io", "sagikazarmark", dag.SetSecret("password", "password")).
		Secret()

	actual, err := secret.Plaintext(ctx)
	if err != nil {
		return err
	}

	const expected = `{"auths":{"docker.io":{"auth":"c2FnaWthemFybWFyazpwYXNzd29yZDI="},"ghcr.io":{"auth":"c2FnaWthemFybWFyazpwYXNzd29yZA=="}}}`

	if actual != expected {
		return fmt.Errorf("secret does not match the expected value\nactual:   %s\nexpected: %s", actual, expected)
	}

	return nil
}

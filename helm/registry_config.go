package main

import (
	"context"
	"crypto/sha1"
	"dagger/helm/internal/dagger"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
)

type RegistryConfig struct {
	// +private
	Auths []RegistryConfigAuth
}

type RegistryConfigAuth struct {
	Address  string
	Username string
	Secret   *dagger.Secret
}

// Add credentials for a registry.
func (m *RegistryConfig) WithRegistryAuth(address string, username string, secret *dagger.Secret) *RegistryConfig {
	m.Auths = append(m.Auths, RegistryConfigAuth{
		Address:  address,
		Username: username,
		Secret:   secret,
	})

	return m
}

// Removes credentials for a registry.
func (m *RegistryConfig) WithoutRegistryAuth(address string) *RegistryConfig {
	m.Auths = slices.DeleteFunc(m.Auths, func(a RegistryConfigAuth) bool {
		return a.Address == address
	})

	return m
}

// Create the registry configuration.
func (m *RegistryConfig) Secret(
	ctx context.Context,

	// Customize the name of the secret.
	//
	// +optional
	name string,
) (*dagger.Secret, error) {
	config, err := m.toConfig(ctx)
	if err != nil {
		return nil, err
	}

	return config.toSecret(name)
}

type RegistryConfigConfig struct {
	Auths map[string]RegistryConfigConfigAuth `json:"auths"`
}

type RegistryConfigConfigAuth struct {
	Auth string `json:"auth"`
}

func (m *RegistryConfig) toConfig(ctx context.Context) (*RegistryConfigConfig, error) {
	config := &RegistryConfigConfig{
		Auths: map[string]RegistryConfigConfigAuth{},
	}

	for _, auth := range m.Auths {
		plaintext, err := auth.Secret.Plaintext(ctx)
		if err != nil {
			return nil, err
		}

		config.Auths[auth.Address] = RegistryConfigConfigAuth{
			Auth: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", auth.Username, plaintext))),
		}
	}

	return config, nil
}

func (c *RegistryConfigConfig) toSecret(name string) (*dagger.Secret, error) {
	out, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	if name == "" {
		h := sha1.New()

		_, err := h.Write(out)
		if err != nil {
			return nil, err
		}

		name = fmt.Sprintf("registry-config-%x", h.Sum(nil))
	}

	return dag.SetSecret(name, string(out)), nil
}

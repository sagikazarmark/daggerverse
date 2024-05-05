// Create an OCI registry configuration file and use it safely with tools, like Helm or Oras.
//
// Tools interacting with an OCI registry usually have their own way to authenticate.
// Helm, for example, provides a command to "login" into a registry, which stores the credentials in a file.
// That is, however, not a safe way to store credentials, especially not in Dagger.
// Credentials persisted in the filesystem make their way into Dagger's layer cache.
//
// This module creates a configuration file and returns it as a Secret that can be mounted safely into a Container.
//
// Be advised that using the tool's built-in authentication mechanism may not work with the configuration file (since it's read only).
//
// You can read more about the topic in [this issue](https://github.com/dagger/dagger/issues/7273).
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type RegistryConfig struct {
	// +private
	Auths []Auth
}

type Auth struct {
	Address  string
	Username string
	Secret   *Secret
}

// Add credentials for a registry.
func (m *RegistryConfig) WithRegistryAuth(address string, username string, secret *Secret) *RegistryConfig {
	m.Auths = append(m.Auths, Auth{
		Address:  address,
		Username: username,
		Secret:   secret,
	})

	return m
}

type Config struct {
	Auths map[string]ConfigAuth `json:"auths"`
}

type ConfigAuth struct {
	Auth string `json:"auth"`
}

// Create the registry configuration.
func (m *RegistryConfig) Secret(
	ctx context.Context,

	// +optional
	// +default="registry-config"
	name string,
) (*Secret, error) {
	config := Config{
		Auths: map[string]ConfigAuth{},
	}

	for _, auth := range m.Auths {
		plaintext, err := auth.Secret.Plaintext(ctx)
		if err != nil {
			return nil, err
		}

		config.Auths[auth.Address] = ConfigAuth{
			Auth: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", auth.Username, plaintext))),
		}
	}

	out, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return dag.SetSecret(name, string(out)), nil
}

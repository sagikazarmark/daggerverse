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
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
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

// Removes credentials for a registry.
func (m *RegistryConfig) WithoutRegistryAuth(address string) *RegistryConfig {
	m.Auths = slices.DeleteFunc(m.Auths, func(a Auth) bool {
		return a.Address == address
	})

	return m
}

// Checks whether the config has any registry credentials.
func (m *RegistryConfig) HasRegistryAuth() bool {
	return len(m.Auths) > 0
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

	// Customize the name of the secret.
	//
	// +optional
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

// MountSecret mounts a registry configuration secret into a container if there is any confuguration in it.
func (m *RegistryConfig) MountSecret(
	ctx context.Context,

	// Container to mount the secret into.
	container *Container,

	// Path to mount the secret into (a common path is ~/.docker/config.json).
	path string,

	// Name of the secret to create and mount.
	//
	// +optional
	secretName string,

	// A user:group to set for the mounted secret.
	//
	// The user and group can either be an ID (1000:1000) or a name (foo:bar).
	//
	// If the group is omitted, it defaults to the same as the user.
	//
	// +optional
	owner string,

	// Permission given to the mounted secret (e.g., 0600).
	//
	// This option requires an owner to be set to be active.
	//
	// +optional
	mode int,
) (*Container, error) {
	if !m.HasRegistryAuth() {
		return container, nil
	}

	secret, err := m.Secret(ctx, secretName)
	if err != nil {
		return nil, err
	}

	return container.WithMountedSecret(path, secret, ContainerWithMountedSecretOpts{
		Owner: owner,
		Mode:  mode,
	}), nil
}

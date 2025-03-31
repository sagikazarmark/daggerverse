package main

import "dagger/vhs/internal/dagger"

// Set an environment variable.
func (m Vhs) WithEnvVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,

	// The value of the environment variable (e.g., "localhost").
	value string,

	// Replace `${VAR}` or $VAR in the value according to the current environment
	// variables defined in the container (e.g., "/opt/bin:$PATH").
	//
	// +optional
	expand bool,
) Vhs {
	m.Container = m.Container.WithEnvVariable(
		name,
		value,
		dagger.ContainerWithEnvVariableOpts{
			Expand: expand,
		},
	)

	return m
}

// Unset an environment variable.
func (m Vhs) WithoutEnvVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,
) Vhs {
	m.Container = m.Container.WithoutEnvVariable(name)

	return m
}

// Set an environment variable containing the given secret.
func (m Vhs) WithSecretVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,

	// The identifier of the secret value.
	secret *dagger.Secret,
) Vhs {
	m.Container = m.Container.WithSecretVariable(name, secret)

	return m
}

// Unset an environment variable containing a secret.
func (m Vhs) WithoutSecretVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,
) Vhs {
	m.Container = m.Container.WithoutSecretVariable(name)

	return m
}

// Set an environment variable.
func (m WithSource) WithEnvVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,

	// The value of the environment variable (e.g., "localhost").
	value string,

	// Replace `${VAR}` or $VAR in the value according to the current environment
	// variables defined in the container (e.g., "/opt/bin:$PATH").
	//
	// +optional
	expand bool,
) WithSource {
	m.Vhs = m.Vhs.WithEnvVariable(name, value, expand)

	return m
}

// Unset an environment variable.
func (m WithSource) WithoutEnvVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,
) WithSource {
	m.Vhs = m.Vhs.WithoutEnvVariable(name)

	return m
}

// Set an environment variable containing the given secret.
func (m WithSource) WithSecretVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,

	// The identifier of the secret value.
	secret *dagger.Secret,
) WithSource {
	m.Vhs = m.Vhs.WithSecretVariable(name, secret)

	return m
}

// Unset an environment variable containing a secret.
func (m WithSource) WithoutSecretVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,
) WithSource {
	m.Vhs = m.Vhs.WithoutSecretVariable(name)

	return m
}

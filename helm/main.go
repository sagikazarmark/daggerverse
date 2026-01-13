// The package manager for Kubernetes.
package main

import (
	"context"
	"dagger/helm/internal/dagger"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "alpine/helm"

type Helm struct {
	Container *dagger.Container

	// +private
	RegistryConfig *dagger.RegistryConfig
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	// +optional
	version string,

	// Custom container to use as a base container.
	// +optional
	container *dagger.Container,
) *Helm {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	container = container.
		// Make sure to run as root for now, so that we know where Helm home is.
		WithUser("root").

		// Do not allow overriding Helm path locations (when using a custom image or container).
		// TODO: add other paths: https://helm.sh/docs/helm/helm/
		WithoutEnvVariable("HELM_HOME").
		WithoutEnvVariable("HELM_REGISTRY_CONFIG")

	// Disable cache mounts for now.
	// Need to figure out if they are needed at all.
	// Note: helm registry auth is stored in ~/.config/helm/registry/config.json (within helm config dir)
	// ctr = ctr.
	// 	WithMountedCache("/root/.cache/helm", dag.CacheVolume("helm-cache")).
	// 	WithMountedCache("/root/.helm", dag.CacheVolume("helm-root")).
	// 	WithMountedCache("/root/.config/helm", dag.CacheVolume("helm-config"))

	return &Helm{
		Container:      container,
		RegistryConfig: dag.RegistryConfig(),
	}
}

// use container for actions that need registry credentials
func (m *Helm) container() *dagger.Container {
	return m.Container.With(func(c *dagger.Container) *dagger.Container {
		return c.WithMountedSecret("/root/.config/helm/registry/config.json", m.RegistryConfig.Secret())
	})

	// return m.Container.With(m.RegistryConfig.SecretMount("/root/.config/helm/registry/config.json").Mount)
}

// Add credentials for a registry.
//
// Note: WithRegistryAuth overrides any previous or subsequent calls to Login/Logout.
//
// +cache="session"
func (m *Helm) WithRegistryAuth(address string, username string, secret *dagger.Secret) *Helm {
	m.RegistryConfig = m.RegistryConfig.WithRegistryAuth(address, username, secret)

	return m
}

// Removes credentials for a registry.
//
// +cache="session"
func (m *Helm) WithoutRegistryAuth(address string) *Helm {
	m.RegistryConfig = m.RegistryConfig.WithoutRegistryAuth(address)

	return m
}

// Mount a file as the kubeconfig file.
func (m *Helm) WithKubeconfigFile(file *dagger.File) *Helm {
	m.Container = m.Container.
		WithoutEnvVariable("KUBECONFIG").
		WithMountedFile("/root/.kube/config", file)

	return m
}

// Mount a secret as the kubeconfig file.
func (m *Helm) WithKubeconfigSecret(secret *dagger.Secret) *Helm {
	m.Container = m.Container.
		WithoutEnvVariable("KUBECONFIG").
		WithMountedSecret("/root/.kube/config", secret)

	return m
}

// Create a new chart directory along with the common files and directories used in a chart.
func (m *Helm) Create(name string) *Chart {
	const workdir = "/work"

	name = path.Clean(name)

	dir := m.Container.
		WithWorkdir(workdir).
		WithExec([]string{"helm", "create", name}).
		Directory(path.Join(workdir, name))

	return m.Chart(dir)
}

// Lint a Helm chart directory.
//
// +cache="session"
func (m *Helm) Lint(
	ctx context.Context,

	// A directory containing a Helm chart.
	chart *dagger.Directory,
) (*dagger.Container, error) {
	chartMetadata, err := getChartMetadata(ctx, chart)
	if err != nil {
		return nil, err
	}

	const workdir = "/work"

	chartName := chartMetadata.Name
	chartPath := path.Join(workdir, chartName)

	args := []string{
		"helm",
		"lint",
		chartPath,
	}

	return m.Container.
		WithWorkdir(workdir).
		WithMountedDirectory(chartPath, chart).WithExec(args), nil
}

// Build a Helm chart package.
func (m *Helm) Package(
	ctx context.Context,

	// A directory containing a Helm chart.
	chart *dagger.Directory,

	// Set the appVersion on the chart to this version.
	//
	// +optional
	appVersion string,

	// Set the version on the chart to this semver version.
	//
	// +optional
	version string,

	// Update dependencies from "Chart.yaml" to dir "charts/" before packaging.
	//
	// +optional
	dependencyUpdate bool,
) (*dagger.File, error) {
	chartMetadata, err := getChartMetadata(ctx, chart)
	if err != nil {
		return nil, err
	}

	const workdir = "/work"

	chartName := chartMetadata.Name
	chartPath := filepath.Join(workdir, chartName)

	args := []string{
		"helm",
		"package",

		"--destination", workdir,
	}

	if appVersion != "" {
		args = append(args, "--app-version", appVersion)
	}

	chartVersion := chartMetadata.Version
	if version != "" {
		args = append(args, "--version", version)
		chartVersion = version
	}

	if dependencyUpdate {
		args = append(args, "--dependency-update")
	}

	args = append(args, chartPath)

	return m.container().
		WithWorkdir(workdir).
		WithMountedDirectory(chartPath, chart).
		WithExec(args).
		File(path.Join(workdir, fmt.Sprintf("%s-%s.tgz", chartName, chartVersion))), nil
}

// Authenticate to an OCI registry.
//
// Note: Login stores credentials in the filesystem in plain text. Use WithRegistryAuth as a safer alternative.
//
// +cache="session"
func (m *Helm) Login(
	ctx context.Context,

	// Host of the OCI registry.
	host string,

	// Registry username.
	username string,

	// Registry password.
	password *dagger.Secret,

	// Allow connections to TLS registry without certs.
	//
	// +optional
	insecure bool,
) *Helm {
	args := []string{
		"helm",
		"registry",
		"login",
		host,
		"--username", username,
		"--password", "$HELM_PASSWORD",
	}

	if insecure {
		args = append(args, "--insecure")
	}

	m.Container = m.Container.
		WithSecretVariable("HELM_PASSWORD", password).
		WithExec([]string{"sh", "-c", strings.Join(args, " ")}).
		WithoutSecretVariable("HELM_PASSWORD")

	return m
}

// Remove credentials stored for an OCI registry.
//
// +cache="session"
func (m *Helm) Logout(host string) *Helm {
	args := []string{
		"registry",
		"logout",
		host,
	}

	m.Container = m.Container.WithExec(args)

	return m
}

// Push a Helm chart package to an OCI registry.
//
// +cache="never"
func (m *Helm) Push(
	ctx context.Context,

	// Packaged Helm chart.
	pkg *dagger.File,

	// OCI registry to push to (including the path except the chart name).
	registry string,

	// Use insecure HTTP connections for the chart upload.
	//
	// +optional
	plainHttp bool,

	// Skip tls certificate checks for the chart upload.
	//
	// +optional
	insecureSkipTlsVerify bool,

	// Verify certificates of HTTPS-enabled servers using this CA bundle.
	//
	// +optional
	caFile *dagger.File,

	// Identify registry client using this SSL certificate file.
	//
	// +optional
	certFile *dagger.File,

	// Identify registry client using this SSL key file.
	//
	// +optional
	keyFile *dagger.Secret,
) error {
	const workdir = "/work"

	chartPath := path.Join(workdir, "chart.tgz")

	container := m.container().
		WithWorkdir(workdir).
		WithMountedFile(chartPath, pkg)

	args := []string{
		"helm",
		"push",

		chartPath,
		registry,
	}

	if plainHttp {
		args = append(args, "--plain-http")
	}

	if insecureSkipTlsVerify {
		args = append(args, "--insecure-skip-tls-verify")
	}

	if caFile != nil {
		container = container.WithMountedFile("/etc/helm/ca.pem", caFile)
		args = append(args, "--ca-file", "/etc/helm/ca.pem")
	}

	if certFile != nil {
		container = container.WithMountedFile("/etc/helm/cert.pem", certFile)
		args = append(args, "--cert-file", "/etc/helm/cert.pem")
	}

	if keyFile != nil {
		container = container.WithMountedSecret("/etc/helm/key.pem", keyFile)
		args = append(args, "--key-file", "/etc/helm/key.pem")
	}

	_, err := container.WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).WithExec(args).Sync(ctx)

	return err
}

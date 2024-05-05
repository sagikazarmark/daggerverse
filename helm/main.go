// The package manager for Kubernetes.
package main

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "alpine/helm"

type Helm struct {
	Container *Container

	// +private
	RegistryConfig *RegistryConfig
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	// +optional
	version string,

	// Custom container to use as a base container.
	// +optional
	container *Container,
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
func (m *Helm) container() *Container {
	return m.Container.
		With(func(c *Container) *Container {
			return m.RegistryConfig.MountSecret(c, "/root/.config/helm/registry/config.json", RegistryConfigMountSecretOpts{
				SecretName: "helm-registry-config",
			})
		})
}

// Add credentials for a registry.
//
// Note: WithRegistryAuth overrides any previous or subsequent calls to Login/Logout.
func (m *Helm) WithRegistryAuth(address string, username string, secret *Secret) *Helm {
	m.RegistryConfig = m.RegistryConfig.WithRegistryAuth(address, username, secret)

	return m
}

// Removes credentials for a registry.
func (m *Helm) WithoutRegistryAuth(address string) *Helm {
	m.RegistryConfig = m.RegistryConfig.WithoutRegistryAuth(address)

	return m
}

// Create a new chart directory along with the common files and directories used in a chart.
func (m *Helm) Create(name string) *Directory {
	const workdir = "/work"

	name = path.Clean(name)

	return m.Container.
		WithWorkdir(workdir).
		WithExec([]string{"create", name}).
		Directory(path.Join(workdir, name))
}

// Lint a Helm chart directory.
func (m *Helm) Lint(
	ctx context.Context,

	// A directory containing a Helm chart.
	chart *Directory,
) (*Container, error) {
	chartMetadata, err := getChartMetadata(ctx, chart)
	if err != nil {
		return nil, err
	}

	const workdir = "/work"

	chartName := chartMetadata.Name
	chartPath := path.Join(workdir, chartName)

	args := []string{
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
	chart *Directory,

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
) (*File, error) {
	chartMetadata, err := getChartMetadata(ctx, chart)
	if err != nil {
		return nil, err
	}

	const workdir = "/work"

	chartName := chartMetadata.Name
	chartPath := filepath.Join(workdir, chartName)

	args := []string{
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
func (m *Helm) Login(
	ctx context.Context,

	// Host of the OCI registry.
	host string,

	// Registry username.
	username string,

	// Registry password.
	password *Secret,

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
		WithExec([]string{"sh", "-c", strings.Join(args, " ")}, ContainerWithExecOpts{SkipEntrypoint: true}).
		WithSecretVariable("HELM_PASSWORD", dag.SetSecret("helm-password-reset", "")) // https://github.com/dagger/dagger/issues/7274

	return m
}

// Remove credentials stored for an OCI registry.
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
func (m *Helm) Push(
	ctx context.Context,

	// Packaged Helm chart.
	pkg *File,

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
	caFile *File,

	// Identify registry client using this SSL certificate file.
	//
	// +optional
	certFile *File,

	// Identify registry client using this SSL key file.
	//
	// +optional
	keyFile *File,
) error {
	const workdir = "/work"

	chartPath := path.Join(workdir, "chart.tgz")

	container := m.container().
		WithWorkdir(workdir).
		WithMountedFile(chartPath, pkg)

	args := []string{
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
		container = container.WithMountedFile("/etc/helm/key.pem", keyFile)
		args = append(args, "--key-file", "/etc/helm/key.pem")
	}

	_, err := container.WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).WithExec(args).Sync(ctx)

	return err
}

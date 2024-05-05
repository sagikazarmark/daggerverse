// The package manager for Kubernetes.
package main

import (
	"context"
	"fmt"
	"path"
	"strings"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "alpine/helm"

type Helm struct {
	Container *Container
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

		// See https://github.com/dagger/dagger/issues/7273
		// WithMountedTemp("/root/.config/helm/registry")

	// Disable cache mounts for now.
	// Need to figure out if they are needed at all.
	// Note: helm registry auth is stored in ~/.config/helm/registry/config.json (within helm config dir)
	// ctr = ctr.
	// 	WithMountedCache("/root/.cache/helm", dag.CacheVolume("helm-cache")).
	// 	WithMountedCache("/root/.helm", dag.CacheVolume("helm-root")).
	// 	WithMountedCache("/root/.config/helm", dag.CacheVolume("helm-config"))

	return &Helm{container}
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
	return m.Chart(chart).Lint(ctx)
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
	pkg, err := m.Chart(chart).Package(ctx, appVersion, version, dependencyUpdate)
	if err != nil {
		return nil, err
	}

	return pkg.File, nil
}

// Authenticate to an OCI registry.
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
	return &Helm{login(m.Container, host, username, password, insecure)}
}

func login(
	ctr *Container,
	host string,
	username string,
	password *Secret,
	insecure bool,
) *Container {
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

	return ctr.
		WithSecretVariable("HELM_PASSWORD", password).
		WithExec([]string{"sh", "-c", strings.Join(args, " ")}, ContainerWithExecOpts{SkipEntrypoint: true}).
		WithSecretVariable("HELM_PASSWORD", dag.SetSecret("helm-password-reset", "")) // https://github.com/dagger/dagger/issues/7274
}

// Remove credentials stored for an OCI registry.
func (m *Helm) Logout(host string) *Helm {
	return &Helm{logout(m.Container, host)}
}

func logout(ctr *Container, host string) *Container {
	args := []string{
		"registry",
		"logout",
		host,
	}

	return ctr.WithExec(args)
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
	p := &Package{
		File:      pkg,
		Container: m.Container,
	}

	return p.Publish(ctx, registry, plainHttp, insecureSkipTlsVerify, caFile, certFile, keyFile)
}

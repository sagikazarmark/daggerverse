// The package manager for Kubernetes.
package main

import (
	"context"
	"fmt"
	"path"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "alpine/helm"

type Helm struct {
	// +private
	Ctr *Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	// +optional
	version string,

	// Custom image reference in "repository:tag" format to use as a base container.
	// +optional
	image string,

	// Custom container to use as a base container.
	// +optional
	container *Container,
) *Helm {
	var ctr *Container

	if version != "" {
		ctr = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	} else if image != "" {
		ctr = dag.Container().From(image)
	} else if container != nil {
		ctr = container
	} else {
		ctr = dag.Container().From(defaultImageRepository)
	}

	// Disable cache mounts for now.
	// Need to figure out if they are needed at all.
	// Note: helm registry auth is stored in ~/.config/helm/registry/config.json (within helm config dir)
	// ctr = ctr..
	// 	// TODO: run as non-root
	// 	WithUser("root").
	// 	WithMountedCache("/root/.cache/helm", dag.CacheVolume("helm-cache")).
	// 	WithMountedCache("/root/.helm", dag.CacheVolume("helm-root")).
	// 	WithMountedCache("/root/.config/helm", dag.CacheVolume("helm-config"))

	return &Helm{ctr}
}

func (m *Helm) Container() *Container {
	return m.Ctr
}

// Create a new chart directory along with the common files and directories used in a chart.
func (m *Helm) Create(name string) *Directory {
	const workdir = "/work"

	name = path.Clean(name)

	return m.Ctr.
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
	// +optional
	appVersion string,

	// Set the version on the chart to this semver version.
	// +optional
	version string,

	// Update dependencies from "Chart.yaml" to dir "charts/" before packaging.
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
	// +optional
	insecure bool,
) (*Helm, error) {
	ctr, err := login(ctx, m.Ctr, host, username, password, insecure)
	if err != nil {
		return nil, err
	}

	return &Helm{ctr}, nil
}

func login(
	ctx context.Context,
	ctr *Container,
	host string,
	username string,
	password *Secret,
	insecure bool,
) (*Container, error) {
	pass, err := password.Plaintext(ctx)
	if err != nil {
		return nil, err
	}

	args := []string{
		"registry",
		"login",
		host,
		"--username", username,
		"--password", pass,
	}

	if insecure {
		args = append(args, "--insecure")
	}

	return ctr.WithExec(args), nil
}

// Remove credentials stored for an OCI registry.
func (m *Helm) Logout(host string) *Helm {
	return &Helm{logout(m.Ctr, host)}
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
	// +optional
	plainHttp bool,

	// Skip tls certificate checks for the chart upload.
	// +optional
	insecureSkipTlsVerify bool,

	// Verify certificates of HTTPS-enabled servers using this CA bundle.
	// +optional
	caFile *File,

	// Identify registry client using this SSL certificate file.
	// +optional
	certFile *File,

	// Identify registry client using this SSL key file.
	// +optional
	keyFile *File,
) error {
	p := &Package{
		File:      pkg,
		Container: m.Ctr,
	}

	return p.Publish(ctx, registry, plainHttp, insecureSkipTlsVerify, caFile, certFile, keyFile)
}

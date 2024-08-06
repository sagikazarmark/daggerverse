package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

// Returns a Helm chart from a source directory.
func (m *Helm) Chart(
	// A directory containing a Helm chart.
	source *dagger.Directory,
) *Chart {
	return &Chart{
		Directory: source,
		Helm:      m,
	}
}

// A Helm chart.
type Chart struct {
	Directory *dagger.Directory

	// +private
	Helm *Helm
}

// Lint a Helm chart.
func (c *Chart) Lint(ctx context.Context) (*dagger.Container, error) {
	return c.Helm.Lint(ctx, c.Directory)
}

// A Helm chart package.
type Package struct {
	File *dagger.File

	// +private
	Helm *Helm

	// +private
	Chart *Chart
}

// Build a Helm chart package.
func (c *Chart) Package(
	ctx context.Context,

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
) (*Package, error) {
	file, err := c.Helm.Package(ctx, c.Directory, appVersion, version, dependencyUpdate)
	if err != nil {
		return nil, err
	}

	return &Package{
		File:  file,
		Helm:  c.Helm,
		Chart: c,
	}, nil
}

// Add credentials for a registry.
func (p *Package) WithRegistryAuth(address string, username string, secret *dagger.Secret) *Package {
	p.Helm = p.Helm.WithRegistryAuth(address, username, secret)

	return p
}

// Removes credentials for a registry.
func (p *Package) WithoutRegistryAuth(address string) *Package {
	p.Helm = p.Helm.WithoutRegistryAuth(address)

	return p
}

// Publishes this Helm chart package to an OCI registry.
func (p *Package) Publish(
	ctx context.Context,

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
	return p.Helm.Push(ctx, p.File, registry, plainHttp, insecureSkipTlsVerify, caFile, certFile, keyFile)
}

package main

import (
	"context"
)

// Returns a Helm chart from a source directory.
func (m *Helm) Chart(
	// A directory containing a Helm chart.
	source *Directory,
) *Chart {
	return &Chart{
		Directory: source,
		Helm:      m,
	}
}

// A Helm chart.
type Chart struct {
	Directory *Directory

	// +private
	Helm *Helm
}

// Lint a Helm chart.
func (m *Chart) Lint(ctx context.Context) (*Container, error) {
	return m.Helm.Lint(ctx, m.Directory)
}

// A Helm chart package.
type Package struct {
	File *File

	// +private
	Helm *Helm
}

// Build a Helm chart package.
func (m *Chart) Package(
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
	file, err := m.Helm.Package(ctx, m.Directory, appVersion, version, dependencyUpdate)
	if err != nil {
		return nil, err
	}

	return &Package{
		File: file,
		Helm: m.Helm,
	}, nil
}

// Add credentials for a registry.
func (m *Package) WithRegistryAuth(address string, username string, secret *Secret) *Package {
	m.Helm = m.Helm.WithRegistryAuth(address, username, secret)

	return m
}

// Removes credentials for a registry.
func (m *Package) WithoutRegistryAuth(address string) *Package {
	m.Helm = m.Helm.WithoutRegistryAuth(address)

	return m
}

// Publishes this Helm chart package to an OCI registry.
func (m *Package) Publish(
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
	return m.Helm.Push(ctx, m.File, registry, plainHttp, insecureSkipTlsVerify, caFile, certFile, keyFile)
}

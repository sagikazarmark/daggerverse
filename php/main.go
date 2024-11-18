// PHP programming language module.

package main

import (
	"dagger/php/internal/dagger"
	"fmt"
	"slices"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "php"

type Php struct {
	// +private
	Ctr *dagger.Container

	// +private
	Extensions []string

	// +private
	ComposerPackages []string

	// Track if composer is installed to avoid installing it unnecessarily.
	//
	// +private
	ComposerInstalled bool

	// Track if docker-php-extension-installer is installed to avoid installing it unnecessarily.
	//
	// +private
	ExtensionInstallerInstalled bool
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	// +default="latest"
	version string,

	// Custom container to use as a base container.
	//
	// +optional
	container *dagger.Container,
) Php {
	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	return Php{
		Ctr: container,
	}
}

// defaultComposerImageRepository is used when no binary is specified.
const defaultComposerImageRepository = "composer"

// TODO: add composer home and cache and path

// Install Composer.
func (m Php) WithComposer(
	// Version (image tag) to use from the official image repository.
	//
	// +optional
	// +default="latest"
	version string,

	// Custom binary to use (takes precedence over version).
	//
	// +optional
	binary *dagger.File,
) Php {
	if binary == nil {
		if version == "" {
			version = "latest"
		}

		binary = dag.Container().
			From(fmt.Sprintf("%s:%s", defaultComposerImageRepository, version)).
			File("/usr/bin/composer")
	}

	m = m.WithExtensions([]string{"bz2", "zip"})

	m.Ctr = m.Ctr.
		WithFile(
			"/usr/local/bin/composer",
			binary,
			dagger.ContainerWithFileOpts{
				Permissions: 0755,
			},
		).
		WithEnvVariable("COMPOSER_ALLOW_SUPERUSER", "1").
		WithEnvVariable("COMPOSER_HOME", "/var/data/composer").
		WithEnvVariable(
			"PATH",
			"${PATH}:/var/data/composer/vendor/bin",
			dagger.ContainerWithEnvVariableOpts{
				Expand: true,
			},
		)
	m.ComposerInstalled = true

	return m
}

func (m Php) withComposer() Php {
	return m.WithComposer("", nil)
}

// Mount a cache volume for Composer cache.
func (m Php) WithComposerCache(
	cache *dagger.CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	//
	// +optional
	source *dagger.Directory,

	// Sharing mode of the cache volume.
	//
	// +optional
	sharing dagger.CacheSharingMode,
) Php {
	m.Ctr = m.Ctr.WithMountedCache(
		"/var/cache/composer",
		cache,
		dagger.ContainerWithMountedCacheOpts{
			Source:  source,
			Sharing: sharing,
		},
	)

	return m
}

// Install a Composer package globally.
func (m Php) WithComposerPackage(name string) Php {
	// make sure the list is copied
	m.ComposerPackages = append(slices.Clone(m.ComposerPackages), name)

	return m
}

// Install a list of Composer packages globally.
func (m Php) WithComposerPackages(name []string) Php {
	// make sure the list is copied
	m.ComposerPackages = append(slices.Clone(m.ComposerPackages), name...)

	return m
}

const (
	// extensionInstallerDownloadUrlTemplate is used to download a specitic version of docker-php-extension-installer.
	extensionInstallerDownloadUrlTemplate = "https://github.com/mlocati/docker-php-extension-installer/releases/download/%s/install-php-extensions"

	latestExtensionInstallerDownloadUrlTemplate = "https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions"
)

// Install docker-php-extension-installer.
func (m Php) WithExtensionInstaller(
	// Version to use from the official repository.
	//
	// +optional
	// +default="latest"
	version string,

	// Custom binary to use (takes precedence over version).
	//
	// +optional
	binary *dagger.File,
) Php {
	if binary == nil {
		if version == "" {
			version = "latest"
		}

		downloadUrl := fmt.Sprintf(extensionInstallerDownloadUrlTemplate, version)
		if version == "latest" {
			downloadUrl = latestExtensionInstallerDownloadUrlTemplate
		}

		binary = dag.HTTP(downloadUrl)
	}

	m.Ctr = m.Ctr.WithFile(
		"/usr/local/bin/install-php-extensions",
		binary,
		dagger.ContainerWithFileOpts{
			Permissions: 0755,
		},
	)
	m.ExtensionInstallerInstalled = true

	return m
}

func (m Php) withExtensionInstaller() Php {
	return m.WithExtensionInstaller("", nil)
}

// Install an extension using docker-php-extension-installer.
func (m Php) WithExtension(name string) Php {
	// make sure the list is copied
	m.Extensions = append(slices.Clone(m.Extensions), name)

	return m
}

// Install a list of extensions using docker-php-extension-installer.
func (m Php) WithExtensions(name []string) Php {
	// make sure the list is copied
	m.Extensions = append(slices.Clone(m.Extensions), name...)

	return m
}

func (m Php) Container() *dagger.Container {
	if len(m.Extensions) > 0 && !m.ExtensionInstallerInstalled {
		m = m.withExtensionInstaller()
	}

	if len(m.ComposerPackages) > 0 && !m.ComposerInstalled {
		m = m.withComposer()
	}

	return m.Ctr.
		With(func(c *dagger.Container) *dagger.Container {
			if len(m.Extensions) > 0 {
				c = c.WithExec(append([]string{"install-php-extensions"}, m.Extensions...))
			}

			return c
		}).
		With(func(c *dagger.Container) *dagger.Container {
			if len(m.ComposerPackages) > 0 {
				c = c.WithExec(append([]string{"composer", "global", "require"}, m.ComposerPackages...))
			}

			return c
		})
}

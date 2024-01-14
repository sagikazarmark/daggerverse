package main

import (
	"fmt"
	"strings"

	"github.com/containerd/containerd/platforms"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "golang"

type Go struct {
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

	// Disable mounting cache volumes.
	// +optional
	disableCache bool,
) *Go {
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

	m := &Go{ctr}

	if !disableCache {
		m = m.
			WithModuleCache(dag.CacheVolume("go-mod"), nil, "").
			WithBuildCache(dag.CacheVolume("go-build"), nil, "")
	}

	return m
}

func (m *Go) Container() *Container {
	return m.Ctr
}

// Set an environment variable.
func (m *Go) WithEnvVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,

	// The value of the environment variable (e.g., "localhost").
	value string,

	// Replace `${VAR}` or $VAR in the value according to the current environment
	// variables defined in the container (e.g., "/opt/bin:$PATH").
	// +optional
	expand bool,
) *Go {
	return &Go{
		m.Ctr.WithEnvVariable(name, value, ContainerWithEnvVariableOpts{
			Expand: expand,
		}),
	}
}

// Set GOOS, GOARCH and GOARM environment variables.
func (m *Go) WithPlatform(
	// Target platform in "[os]/[platform]/[version]" format (e.g., "darwin/arm64/v7", "windows/amd64", "linux/arm64").
	platform Platform,
) *Go {
	if platform == "" {
		return m
	}

	p := platforms.MustParse(string(platform))

	ctr := m.Ctr.
		WithEnvVariable("GOOS", p.OS).
		WithEnvVariable("GOARCH", p.Architecture)

	if p.Variant != "" {
		ctr = ctr.WithEnvVariable("GOARM", p.Variant)
	}

	return &Go{ctr}
}

// Set CGO_ENABLED environment variable to 1.
func (m *Go) WithCgoEnabled() *Go {
	return &Go{m.Ctr.WithEnvVariable("CGO_ENABLED", "1")}
}

// Set CGO_ENABLED environment variable to 0.
func (m *Go) WithCgoDisabled() *Go {
	return &Go{m.Ctr.WithEnvVariable("CGO_ENABLED", "0")}
}

// Mount a cache volume for Go module cache.
func (m *Go) WithModuleCache(
	cache *CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	// +optional
	source *Directory,

	// Sharing mode of the cache volume.
	// +optional
	sharing CacheSharingMode,
) *Go {
	return &Go{m.Ctr.WithMountedCache("/go/pkg/mod", cache, ContainerWithMountedCacheOpts{
		Source:  source,
		Sharing: sharing,
	})}
}

// Mount a cache volume for Go build cache.
func (m *Go) WithBuildCache(
	cache *CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	// +optional
	source *Directory,

	// Sharing mode of the cache volume.
	// +optional
	sharing CacheSharingMode,
) *Go {
	return &Go{m.Ctr.WithMountedCache("/root/.cache/go-build", cache, ContainerWithMountedCacheOpts{
		Source:  source,
		Sharing: sharing,
	})}
}

// Run a Go command.
func (m *Go) Exec(
	// Arguments to pass to the Go command.
	args []string,

	// Source directory to mount.
	// +optional
	src *Directory,

	// Target platform in "[os]/[platform]/[version]" format (e.g., "darwin/arm64/v7", "windows/amd64", "linux/arm64").
	// +optional
	platform Platform,
) *Container {
	if platform != "" {
		m = m.WithPlatform(platform)
	}

	if src != nil {
		return m.WithSource(src).Exec(args, Platform(""))
	}

	return m.Ctr.WithExec(args)
}

// Build a binary.
func (m *Go) Build(
	// Source directory to mount.
	source *Directory,

	// Package to compile.
	// +optional
	pkg string,

	// A list of additional build tags to consider satisfied during the build.
	// +optional
	tags []string,

	// Remove all file system paths from the resulting executable.
	// +optional
	trimpath bool,

	// Additional args to pass to the build command.
	// +optional
	rawArgs []string,

	// Target platform in "[os]/[platform]/[version]" format (e.g., "darwin/arm64/v7", "windows/amd64", "linux/arm64").
	// +optional
	platform Platform,
) *File {
	return m.WithSource(source).Build(pkg, tags, trimpath, rawArgs, platform)
}

// Mount a source directory.
func (m *Go) WithSource(
	// Source directory to mount.
	source *Directory,
) *WithSource {
	const workdir = "/work"

	return &WithSource{
		&Go{
			m.Ctr.
				WithWorkdir(workdir).
				WithMountedDirectory(workdir, source),
		},
	}
}

type WithSource struct {
	// +private
	Go *Go
}

func (m *WithSource) Container() *Container {
	return m.Go.Ctr
}

// Set an environment variable.
func (m *WithSource) WithEnvVariable(
	// The name of the environment variable (e.g., "HOST").
	name string,

	// The value of the environment variable (e.g., "localhost").
	value string,

	// Replace `${VAR}` or $VAR in the value according to the current environment
	// variables defined in the container (e.g., "/opt/bin:$PATH").
	// +optional
	expand bool,
) *WithSource {
	return &WithSource{m.Go.WithEnvVariable(name, value, expand)}
}

// Set GOOS, GOARCH and GOARM environment variables.
func (m *WithSource) WithPlatform(
	// Target platform in "[os]/[platform]/[version]" format (e.g., "darwin/arm64/v7", "windows/amd64", "linux/arm64").
	platform Platform,
) *WithSource {
	return &WithSource{m.Go.WithPlatform(platform)}
}

// Set CGO_ENABLED environment variable to 1.
func (m *WithSource) WithCgoEnabled() *WithSource {
	return &WithSource{m.Go.WithCgoEnabled()}
}

// Set CGO_ENABLED environment variable to 0.
func (m *WithSource) WithCgoDisabled() *WithSource {
	return &WithSource{m.Go.WithCgoDisabled()}
}

// Mount a cache volume for Go module cache.
func (m *WithSource) WithModuleCache(
	cache *CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	// +optional
	source *Directory,

	// Sharing mode of the cache volume.
	// +optional
	sharing CacheSharingMode,
) *WithSource {
	return &WithSource{m.Go.WithModuleCache(cache, source, sharing)}
}

// Mount a cache volume for Go build cache.
func (m *WithSource) WithBuildCache(
	cache *CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	// +optional
	source *Directory,

	// Sharing mode of the cache volume.
	// +optional
	sharing CacheSharingMode,
) *WithSource {
	return &WithSource{m.Go.WithBuildCache(cache, source, sharing)}
}

// Run a Go command.
func (m *WithSource) Exec(
	// Arguments to pass to the Go command.
	args []string,

	// Target platform in "[os]/[platform]/[version]" format (e.g., "darwin/arm64/v7", "windows/amd64", "linux/arm64").
	// +optional
	platform Platform,
) *Container {
	if platform != "" {
		m = m.WithPlatform(platform)
	}

	return m.Go.Ctr.WithExec(args)
}

// Compile the packages into a binary.
func (m *WithSource) Build(
	// Package to compile.
	// +optional
	pkg string,

	// A list of additional build tags to consider satisfied during the build.
	// +optional
	tags []string,

	// Remove all file system paths from the resulting executable.
	// +optional
	trimpath bool,

	// Additional args to pass to the build command.
	// +optional
	rawArgs []string,

	// Target platform in "[os]/[platform]/[version]" format (e.g., "darwin/arm64/v7", "windows/amd64", "linux/arm64").
	// +optional
	platform Platform,
) *File {
	binaryPath := "/out/binary"

	args := []string{"go", "build", "-o", binaryPath}

	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}

	if trimpath {
		args = append(args, "-trimpath")
	}

	if len(rawArgs) > 0 {
		args = append(args, rawArgs...)
	}

	if pkg != "" {
		args = append(args, pkg)
	}

	return m.Exec(args, platform).File(binaryPath)
}

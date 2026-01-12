// Rust programming language module.
package main

import (
	"context"
	"dagger/rust/internal/dagger"
	"fmt"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "rust"

type Rust struct {
	Source *dagger.Directory

	// +private
	Container *dagger.Container
}

func New(
	ctx context.Context,

	// Source directory.
	source *dagger.Directory,

	// Components to install.
	//
	// +optional
	components []string,

	// Targets to install.
	//
	// +optional
	targets []string,

	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	version string,

	// Custom container to use as a base container.
	//
	// +optional
	container *dagger.Container,

	// Disable mounting cache volumes.
	//
	// +optional
	disableCache bool,
) (*Rust, error) {
	exists, err := source.Exists(ctx, "rust-toolchain.toml")
	if err != nil {
		return nil, err
	}

	var toolchain struct {
		Channel    string
		Components []string
		Targets    []string

		// TODO: add profile?
	}

	if exists {
		var file struct {
			Toolchain struct {
				Channel    string
				Components []string
				Targets    []string

				// TODO: add profile?
			} `toml:"toolchain"`
		}

		content, err := source.File("rust-toolchain.toml").Contents(ctx)
		if err != nil {
			return nil, err
		}

		if err := toml.Unmarshal([]byte(content), &file); err != nil {
			return nil, err
		}

		toolchain = file.Toolchain

		if version == "" {
			version = toolchain.Channel

			// if strings.HasPrefix(version, "nightly-") {
			// 	version = "nightly"
			// }
		}
	}

	if container == nil {
		if version == "" {
			version = "latest"
		}

		container = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	// if strings.HasPrefix(toolchain.Channel, "nightly-") {
	// 	container = container.WithExec([]string{"rustup", "override", "set", toolchain.Channel})
	// }

	if components = append(components, toolchain.Components...); len(components) > 0 {
		for _, component := range components {
			container = container.WithExec([]string{"rustup", "component", "add", component})
		}
	}

	if targets = append(targets, toolchain.Targets...); len(targets) > 0 {
		for _, target := range targets {
			container = container.WithExec([]string{"rustup", "target", "add", target})
		}
	}

	container = container.WithWorkdir("/work/src")

	m := &Rust{
		Source:    source,
		Container: container,
	}

	if !disableCache {
		m = m.
			WithRegistryCache(dag.CacheVolume("rust-registry"), nil, "").
			WithGitCache(dag.CacheVolume("rust-git"), nil, "")
		// TODO: add target caching?
	}

	return m, nil
}

// Mount a cache volume for Cargo registry cache.
func (m *Rust) WithRegistryCache(
	cache *dagger.CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	//
	// +optional
	source *dagger.Directory,

	// Sharing mode of the cache volume.
	//
	// +optional
	sharing dagger.CacheSharingMode,
) *Rust {
	m.Container = m.Container.WithMountedCache(
		"/root/.cargo/registry",
		cache,
		dagger.ContainerWithMountedCacheOpts{
			Source:  source,
			Sharing: sharing,
		},
	)

	return m
}

// Mount a cache volume for Cargo git cache.
func (m *Rust) WithGitCache(
	cache *dagger.CacheVolume,

	// Identifier of the directory to use as the cache volume's root.
	//
	// +optional
	source *dagger.Directory,

	// Sharing mode of the cache volume.
	//
	// +optional
	sharing dagger.CacheSharingMode,
) *Rust {
	m.Container = m.Container.WithMountedCache(
		"/root/.cache/git",
		cache,
		dagger.ContainerWithMountedCacheOpts{
			Source:  source,
			Sharing: sharing,
		},
	)

	return m
}

func (m *Rust) container() *dagger.Container {
	return m.Container.WithMountedDirectory(".", m.Source)
}

// Mount a cache volume for Cargo registry cache.
func (m *Rust) WithChef(
	// Version of cargo-chef to install.
	//
	// +optional
	version string,
) *Rust {
	pkg := "cargo-chef"
	if version != "" {
		pkg += "@" + version
	}

	recipe := m.Container.
		WithExec([]string{"cargo", "install", pkg}).
		WithMountedDirectory(".", m.Source).
		WithExec([]string{"cargo", "chef", "prepare"}).
		File("recipe.json")

	m.Container = m.Container.
		WithMountedFile("/work/recipe.json", recipe).
		WithExec([]string{"cargo", "chef", "prepare", "--recipe-path", "/work/recipe.json"})

	return m
}

func (m *Rust) cargo() []string {
	return []string{"cargo", "--locked"}
}

type cargoBuiltin struct {
	// Package Selection

	// Package to build.
	//
	// +optional
	pkg string

	// Target Selection (TODO)

	// Feature Selection

	// List of features to activate.
	//
	// +optional
	features []string

	// Activate all available features.
	//
	// +optional
	allFeatures bool

	// Do not activate the `default` feature.
	//
	// +optional
	noDefaultFeatures bool

	// Compilation options

	// Build artifacts in release mode, with optimizations.
	//
	// +optional
	release bool

	// Build artifacts with the specified profile.
	//
	// +optional
	profile string

	// Build for the target triple.
	//
	// +optional
	target string
}

func (a cargoBuiltin) args() []string {
	args := []string{}

	if a.pkg != "" {
		args = append(args, "-p", a.pkg)
	}

	if len(a.features) > 0 {
		args = append(args, "--features", strings.Join(a.features, ","))
	}
	if a.allFeatures {
		args = append(args, "--all-features")
	}
	if a.noDefaultFeatures {
		args = append(args, "--no-default-features")
	}

	if a.release {
		args = append(args, "--release")
	}

	if a.profile != "" {
		args = append(args, "--profile", a.profile)
	}

	if a.target != "" {
		args = append(args, "--target", a.target)
	}

	return args
}

// Compile a local package and all of its dependencies.
func (m *Rust) Build(
	// Package Selection

	// Package to build.
	//
	// +optional
	pkg string,

	// Target Selection (TODO)

	// Feature Selection

	// List of features to activate.
	//
	// +optional
	features []string,

	// Activate all available features.
	//
	// +optional
	allFeatures bool,

	// Do not activate the `default` feature.
	//
	// +optional
	noDefaultFeatures bool,

	// Compilation options

	// Build artifacts in release mode, with optimizations.
	//
	// +optional
	release bool,

	// Build artifacts with the specified profile.
	//
	// +optional
	profile string,

	// Build for the target triple.
	//
	// +optional
	target string,
) *dagger.Directory {
	args := append(m.cargo(), "build")

	builtin := cargoBuiltin{
		pkg:               pkg,
		features:          features,
		allFeatures:       allFeatures,
		noDefaultFeatures: noDefaultFeatures,
		release:           release,
		profile:           profile,
	}

	args = append(args, builtin.args()...)

	return m.container().WithExec(args).Directory("./target")
}

// Execute all unit and integration tests and build examples of a local package.
func (m *Rust) Test(
	// If specified, only run tests containing this string in their names.
	//
	// +optional
	testName string,

	// Arguments for the test binary.
	//
	// +optional
	args []string,

	// Package Selection

	// Package to build.
	//
	// +optional
	pkg string,

	// Target Selection (TODO)

	// Feature Selection

	// List of features to activate.
	//
	// +optional
	features []string,

	// Activate all available features.
	//
	// +optional
	allFeatures bool,

	// Do not activate the `default` feature.
	//
	// +optional
	noDefaultFeatures bool,

	// Compilation options

	// Build artifacts in release mode, with optimizations.
	//
	// +optional
	release bool,

	// Build artifacts with the specified profile.
	//
	// +optional
	profile string,

	// Build for the target triple.
	//
	// +optional
	target string,
) *dagger.Container {
	rawArgs := args

	args = append(m.cargo(), "test")

	if testName != "" {
		args = append(args, testName)
	}

	builtin := cargoBuiltin{
		pkg:               pkg,
		features:          features,
		allFeatures:       allFeatures,
		noDefaultFeatures: noDefaultFeatures,
		release:           release,
		profile:           profile,
	}

	args = append(args, builtin.args()...)

	if len(rawArgs) > 0 {
		args = append(args, "--")
		args = append(args, rawArgs...)
	}

	return m.container().WithExec(args)
}

func (m *Rust) Check(
	// Package Selection

	// Package to build.
	//
	// +optional
	pkg string,

	// Target Selection (TODO)

	// Feature Selection

	// List of features to activate.
	//
	// +optional
	features []string,

	// Activate all available features.
	//
	// +optional
	allFeatures bool,

	// Do not activate the `default` feature.
	//
	// +optional
	noDefaultFeatures bool,

	// Compilation options

	// Build artifacts in release mode, with optimizations.
	//
	// +optional
	release bool,

	// Build artifacts with the specified profile.
	//
	// +optional
	profile string,
) *dagger.Container {
	args := append(m.cargo(), "check")

	builtin := cargoBuiltin{
		pkg:               pkg,
		features:          features,
		allFeatures:       allFeatures,
		noDefaultFeatures: noDefaultFeatures,
		release:           release,
		profile:           profile,
	}

	args = append(args, builtin.args()...)

	return m.container().WithExec(args)
}

type Format struct {
	// +private
	Rust *Rust

	// +private
	Pkg string
}

func (m *Format) args() []string {
	args := []string{}

	if m.Pkg != "" {
		args = append(args, "--package", m.Pkg)
	}

	return args
}

// This utility formats all bin and lib files of the current crate using rustfmt.
func (m *Rust) Format(
	// Specify package to format.
	//
	// +optional
	pkg string,
) *Format {
	return &Format{
		Rust: m,
		Pkg:  pkg,
	}
}

// Format code.
func (m *Format) Run() *dagger.Changeset {
	args := append(m.Rust.cargo(), "fmt")

	if fmtArgs := m.args(); len(fmtArgs) > 0 {
		args = append(args, "--")
		args = append(args, fmtArgs...)
	}

	return m.Rust.container().WithExec(args).Directory(".").Changes(m.Rust.Source)
}

// Run in check mode.
func (m *Format) Check() *dagger.Container {
	args := append(m.Rust.cargo(), "fmt")

	if fmtArgs := m.args(); len(fmtArgs) > 0 {
		args = append(args, "--")
		args = append(args, fmtArgs...)
	}

	return m.Rust.container().WithExec(args)
}

type Clippy struct {
	// +private
	Rust *Rust
}

// Checks a package to catch common mistakes and improve your Rust code.
func (m *Rust) Clippy() *Clippy {
	return &Clippy{
		Rust: m,
	}
}

// Run checks.
func (m *Clippy) Run(
	// Run Clippy only on the given crate, without linting the dependencies.
	//
	// +optional
	noDeps bool,
) *dagger.Container {
	args := append(m.Rust.cargo(), "clippy")

	if noDeps {
		args = append(args, "--no-deps")
	}

	return m.Rust.container().WithExec(args)
}

// Fix checks.
func (m *Clippy) Fix() *dagger.Changeset {
	args := append(m.Rust.cargo(), "clippy", "--fix")

	return m.Rust.container().WithExec(args).Directory(".").Changes(m.Rust.Source)
}

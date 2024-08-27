// Build Caddy with plugins.
package main

import (
	"dagger/xcaddy/internal/dagger"
	"fmt"
	"runtime"
	"strings"

	"github.com/containerd/platforms"
	mod "golang.org/x/mod/module"
)

// defaultGoImageRepository is used when no image is specified.
const defaultGoImageRepository = "golang"

type Xcaddy struct {
	// +private
	Container *dagger.Container
}

func New(
	// xcaddy version.
	//
	// +optional
	version string,

	// Custom container to use as an xcaddy (and Go) base container.
	//
	// +optional
	container *dagger.Container,

	// Version (image tag) to use from the official image repository as a Go base container.
	//
	// +optional
	goVersion string,

	// Custom container to use as a Go base container.
	//
	// +optional
	goContainer *dagger.Container,
) *Xcaddy {
	if container == nil {
		if goContainer == nil {
			if goVersion == "" {
				goVersion = "latest"
			}

			goContainer = dag.Container().From(fmt.Sprintf("%s:%s", defaultGoImageRepository, goVersion))
		}

		var binary *dagger.File

		if version == "" {
			binary = goContainer.
				WithEnvVariable("GOBIN", "/work").
				WithExec([]string{"go", "install", "github.com/caddyserver/xcaddy/cmd/xcaddy@latest"}).
				File("/work/xcaddy")
		} else {
			fileName := fmt.Sprintf("xcaddy_%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
			binary = dag.Arc().Unarchive(
				dag.HTTP(
					fmt.Sprintf("https://github.com/caddyserver/xcaddy/releases/download/v%s/%s", version, fileName),
				).WithName(fileName),
			).File("xcaddy_0/xcaddy")
		}

		container = goContainer.
			With(resetEnvVariables).
			WithFile("/usr/local/bin/xcaddy", binary)
	} else {
		container = resetEnvVariables(container)
	}

	return &Xcaddy{
		Container: container,
	}
}

func resetEnvVariables(c *dagger.Container) *dagger.Container {
	return c.
		WithoutEnvVariable("CADDY_VERSION").
		WithoutEnvVariable("XCADDY_SKIP_BUILD").
		WithoutEnvVariable("XCADDY_SKIP_CLEANUP")
}

// Build Caddy with plugins.
func (m *Xcaddy) Build(
	// Caddy version.
	//
	// +optional
	version string,

	// Enables the Go race detector in the build.
	//
	// +optional
	race bool,

	// Enables the DWARF debug information in the build.
	//
	// +optional
	debug bool,
) *Build {
	return &Build{
		Version: version,
		Race:    race,
		Debug:   debug,

		Xcaddy: m,
	}
}

// Build Caddy with plugins.
type Build struct {
	// Caddy version.
	Version string

	// Whether the Go race detector is enabled in the build.
	Race bool

	// Whether the DWARF debug information is enabled in the build.
	Debug bool

	// List of plugins to include.
	Plugins []GoModule

	// List of modules to replace.
	Replacements []GoModule

	// List of embedded directories.
	Embeds []Embed

	// +private
	Xcaddy *Xcaddy
}

type GoModule struct {
	// Go module path.
	Path string

	// Go module version (optional).
	Version string

	// Local replacement directory (optional).
	Replacement *dagger.Directory
}

type Embed struct {
	// Name of the embedded directory.
	Alias string

	// Directory to embed in the binary.
	Directory *dagger.Directory
}

// Add plugins to the Caddy build.
func (b *Build) Plugin(
	// Go module path.
	module string,

	// Go module version.
	//
	// +optional
	version string,

	// Local replacement directory.
	//
	// +optional
	replacement *dagger.Directory,
) (*Build, error) {
	// This is to make sure there is no path magic applied to the module to escape the working directory.
	if err := mod.CheckPath(module); err != nil {
		return nil, err
	}

	b.Plugins = append(b.Plugins, GoModule{
		Path:        module,
		Version:     version,
		Replacement: replacement,
	})

	return b, nil
}

// Replace Caddy dependencies.
func (b *Build) Replace(
	// Go module path.
	module string,

	// Go module version.
	//
	// +optional
	version string,

	// Local replacement directory.
	//
	// +optional
	replacement *dagger.Directory,
) (*Build, error) {
	// This is to make sure there is no path magic applied to the module to escape the working directory.
	if err := mod.CheckPath(module); err != nil {
		return nil, err
	}

	b.Replacements = append(b.Replacements, GoModule{
		Path:        module,
		Version:     version,
		Replacement: replacement,
	})

	return b, nil
}

// Embed a directory in the Caddy binary.
func (b *Build) Embed(
	// Name of the embedded directory.
	alias string,

	// Directory to embed in the binary.
	directory *dagger.Directory,
) (*Build, error) {
	if strings.Contains(alias, "/") {
		return nil, fmt.Errorf("alias cannot contain a slash")
	}

	b.Embeds = append(b.Embeds, Embed{
		Alias:     alias,
		Directory: directory,
	})

	return b, nil
}

func (b *Build) build(
	platform dagger.Platform,
	skipBuild bool,
) *dagger.Container {
	container := b.Xcaddy.Container.
		WithWorkdir("/work").
		With(func(c *dagger.Container) *dagger.Container {
			if platform == "" {
				return c
			}

			p := platforms.MustParse(string(platform))

			c = c.
				WithEnvVariable("GOOS", p.OS).
				WithEnvVariable("GOARCH", p.Architecture)

			if p.Variant != "" {
				return c.WithEnvVariable("GOARM", p.Variant)
			}

			return c
		})

	args := []string{"xcaddy", "build"}

	if b.Version != "" {
		args = append(args, b.Version)
	}

	container, args = appendGoModules(container, args, "with", b.Plugins)
	container, args = appendGoModules(container, args, "replace", b.Replacements)

	for _, embed := range b.Embeds {
		mountPath := "/work/embed/" + embed.Alias

		container = container.WithMountedDirectory(mountPath, embed.Directory)
		args = append(args, "--embed", embed.Alias+":"+mountPath)
	}

	return container.
		With(func(c *dagger.Container) *dagger.Container {
			if b.Race {
				c = c.WithEnvVariable("XCADDY_RACE_DETECTOR", "1")
			}

			if b.Debug {
				c = c.WithEnvVariable("XCADDY_DEBUG", "1")
			}

			if skipBuild {
				c = c.WithEnvVariable("XCADDY_SKIP_BUILD", "1")
			}

			return c
		}).
		WithExec(args)
}

func appendGoModules(container *dagger.Container, args []string, kind string, modules []GoModule) (*dagger.Container, []string) {
	for _, module := range modules {
		arg := module.Path

		if module.Version != "" {
			arg += "@" + module.Version
		}

		if module.Replacement != nil {
			mountPath := fmt.Sprintf("/work/%s/%s", kind, module.Path)

			container = container.WithMountedDirectory(mountPath, module.Replacement)

			arg += "=" + mountPath
		}

		args = append(args, "--"+kind, arg)
	}

	return container, args
}

// Return a Caddy binary.
func (b *Build) Binary(
	// Target platform in "[os]/[platform]/[version]" format (e.g., "darwin/arm64/v7", "windows/amd64", "linux/arm64").
	//
	// +optional
	platform dagger.Platform,
) *dagger.File {
	return b.build(platform, false).File("/work/caddy")
}

// Return a Caddy container.
func (b *Build) Container(
	// Use the specified base image.
	//
	// +optional
	// +default="caddy"
	base string,

	// Target platform in "[os]/[platform]/[version]" format (e.g., "darwin/arm64/v7", "windows/amd64", "linux/arm64").
	//
	// +optional
	platform dagger.Platform,
) *dagger.Container {
	var opts dagger.ContainerOpts

	if platform != "" {
		opts.Platform = platform
	}

	return dag.Container(opts).
		From(base).
		WithFile("/usr/bin/caddy", b.Binary(platform))
}

// Open a terminal to inspect the build files.
func (b *Build) Inspect() *dagger.Container {
	return b.build(dagger.Platform(""), true).Terminal()
}

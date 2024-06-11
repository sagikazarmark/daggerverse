// Build Caddy with plugins.
package main

import (
	"fmt"
	"runtime"

	"github.com/containerd/containerd/platforms"
)

type Xcaddy struct {
	// xcaddy version.
	Version string

	// +private
	GoVersion string
}

func New(
	// xcaddy version.
	//
	// +optional
	// +default="0.4.2"
	version string,

	// Go version.
	//
	// +optional
	goVersion string,
) *Xcaddy {
	return &Xcaddy{
		Version:   version,
		GoVersion: goVersion,
	}
}

// Build builds Caddy with plugins.
func (m *Xcaddy) Build(
	// Caddy version.
	//
	// +optional
	version string,
) *Build {
	return &Build{
		Version: version,

		Xcaddy: m,
	}
}

type Build struct {
	Version string

	WithModules []WithModule

	// +private
	Xcaddy *Xcaddy
}

func (b *Build) WithVersion(
	// Caddy version.
	version string,
) *Build {
	b.Version = version

	return b
}

func (b *Build) WithModule(
	module string,

	// +optional
	version string,

	// +optional
	replacement *Directory,
) *Build {
	b.WithModules = append(b.WithModules, WithModule{
		Module:      module,
		Version:     version,
		Replacement: replacement,
	})

	return b
}

type WithModule struct {
	Module      string
	Version     string
	Replacement *Directory
}

func (b *Build) build(
	platform Platform,
	skipBuild bool,
) *Container {
	name := fmt.Sprintf("xcaddy_%s_%s_%s.tar.gz", b.Xcaddy.Version, runtime.GOOS, runtime.GOARCH)
	binary := dag.Arc().Unarchive(
		dag.HTTP(
			fmt.Sprintf("https://github.com/caddyserver/xcaddy/releases/download/v%s/xcaddy_%s_%s_%s.tar.gz", b.Xcaddy.Version, b.Xcaddy.Version, runtime.GOOS, runtime.GOARCH),
		).WithName(name),
	).File("xcaddy_0/xcaddy")

	// if platform == "" {
	// 	p, err := dag.DefaultPlatform(ctx)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	platform = p
	// }

	goImage := "golang"

	if b.Xcaddy.GoVersion != "" {
		goImage += ":" + b.Xcaddy.GoVersion
	}

	container := dag.Container().
		From(goImage).
		WithFile("/usr/local/bin/xcaddy", binary).
		WithWorkdir("/work").
		With(func(c *Container) *Container {
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

	for _, withModule := range b.WithModules {
		arg := withModule.Module

		if withModule.Version != "" {
			arg += "@" + withModule.Version
		}

		if withModule.Replacement != nil {
			mountPath := "/work/with/" + withModule.Module

			container = container.WithMountedDirectory(mountPath, withModule.Replacement)

			arg += "=" + mountPath
		}

		args = append(args, "--with", arg)
	}

	return container.
		With(func(c *Container) *Container {
			if skipBuild {
				c = c.WithEnvVariable("XCADDY_SKIP_BUILD", "1")
			}

			return c
		}).
		WithExec(args)
}

func (b *Build) Binary(
	// +optional
	platform Platform,
) *File {
	return b.build(platform, false).File("/work/caddy")
}

func (b *Build) Inspect(
	// +optional
	platform Platform,
) *Terminal {
	return b.build(platform, true).Terminal()
}

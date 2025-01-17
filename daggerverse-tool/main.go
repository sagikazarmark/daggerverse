// Manage your Daggerverse repository with ease.

package main

import (
	"context"
	"dagger/daggerverse-tool/internal/dagger"
	_ "embed"
	"fmt"

	"github.com/containerd/platforms"
)

// TODO: Flake hacks
// TODO: Go SDK hacks

type DaggerverseTool struct {
	// Daggerverse source directory
	//
	// +private
	Source *dagger.Directory
}

func New(
	// Project source directory.
	//
	// +optional
	source *dagger.Directory,
) *DaggerverseTool {
	return &DaggerverseTool{
		Source: source,
	}
}

//go:embed templates/tests/main.go
var testTemplate string

// Initialize a new module.
func (m *DaggerverseTool) Init(
	ctx context.Context,

	name string,
) (*dagger.Directory, error) {
	daggerBinary, err := daggerPrebuiltBinary(ctx)
	if err != nil {
		return nil, err
	}

	base := dag.Container().
		From("alpine").
		WithExec([]string{"apk", "add", "jq", "moreutils", "go", "git"}).
		WithFile("/usr/local/bin/dagger", daggerBinary)

	return base.
		WithWorkdir("/work").
		With(func(c *dagger.Container) *dagger.Container {
			// Use existing source (if any)
			if m.Source != nil {
				return c.WithMountedDirectory("", m.Source)
			}

			// Create fake source otherwise
			return c.
				WithExec([]string{"git", "init"}). // Trick Dagger into accepting this dir as a root
				WithNewFile("LICENSE", "")         // Trick Dagger into not creating a LICENSE file
		}).
		WithWorkdir("/work/module").
		With(daggerExec([]string{"init", "--sdk", "go", "--source", ".", "--name", name})).
		WithExec([]string{
			"sh", "-c",
			`jq '.exclude = ["../.direnv", "../.devenv", "../go.work", "../go.work.sum"]' dagger.json | sponge dagger.json`,
		}).
		With(daggerExec([]string{"develop"})).
		WithWorkdir("/work/module/tests").
		With(daggerExec([]string{"init", "--sdk", "go", "--source", ".", "--name", "tests"})).
		WithExec([]string{
			"sh", "-c",
			`jq '.exclude = ["../../.direnv", "../../.devenv", "../../go.work", "../../go.work.sum"]' dagger.json | sponge dagger.json`,
		}).
		WithExec([]string{"go", "mod", "edit", "-module", fmt.Sprintf("dagger/%s/tests", name), "go.mod"}).
		WithNewFile("main.go", testTemplate).
		With(daggerExec([]string{"develop"})).
		WithWorkdir("/work/module").
		Directory(""), nil
}

func daggerSourceBinary() *dagger.File {
	return dag.DaggerCli().Binary()
}

func daggerPrebuiltBinary(ctx context.Context) (*dagger.File, error) {
	version, err := dag.Version(ctx)
	if err != nil {
		return nil, err
	}

	rawPlatform, err := dag.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	platform := platforms.MustParse(string(rawPlatform))

	archive := dag.HTTP(fmt.Sprintf("https://github.com/dagger/dagger/releases/download/%s/dagger_%s_%s_%s.tar.gz", version, version, platform.OS, platform.Architecture)).WithName("dagger.tar.gz")

	return dag.Arc().Unarchive(archive).File("dagger/dagger"), nil
}

func daggerExec(args []string) func(*dagger.Container) *dagger.Container {
	return func(c *dagger.Container) *dagger.Container {
		return c.WithExec(append([]string{"dagger"}, args...), dagger.ContainerWithExecOpts{ExperimentalPrivilegedNesting: true})
	}
}

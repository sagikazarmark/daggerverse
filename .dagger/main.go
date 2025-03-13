package main

import (
	"dagger/daggerverse/internal/dagger"
)

type Daggerverse struct {
	// Project source directory
	//
	// +private
	Source *dagger.Directory
}

func New(
	// Project source directory.
	//
	// +defaultPath="/"
	// +ignore=[".devenv", ".direnv", ".github", "go.work", "go.work.sum"]
	source *dagger.Directory,
) *Daggerverse {
	return &Daggerverse{
		Source: source,
	}
}

// Initialize a new module.
func (m *Daggerverse) Init(name string) *dagger.Directory {
	return dag.DaggerverseTool().Init(name)
}

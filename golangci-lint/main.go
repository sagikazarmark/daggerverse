package main

import (
	"fmt"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "golangci/golangci-lint"

type GolangciLint struct{}

func (m *GolangciLint) Run(
	version Optional[string],
	image Optional[string],
	container Optional[*Container],
	goVersion Optional[string],
	goImage Optional[string],
	goContainer Optional[*Container],
	source Optional[*Directory],
	verbose Optional[bool],
	timeout Optional[string],
) *Container {
	var golangciLint *Container

	if v, ok := version.Get(); ok {
		golangciLint = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, v))
	} else if i, ok := image.Get(); ok {
		golangciLint = dag.Container().From(i)
	} else if c, ok := container.Get(); ok {
		golangciLint = c
	} else {
		golangciLint = dag.Container().From(defaultImageRepository)
	}

	var goBase *GoBase

	if v, ok := goVersion.Get(); ok {
		goBase = dag.Go().FromVersion(v)
	} else if i, ok := image.Get(); ok {
		goBase = dag.Go().FromImage(i)
	} else if c, ok := container.Get(); ok {
		goBase = dag.Go().FromContainer(c)
	} else {
		goBase = dag.Go().FromVersion("latest")
	}

	args := []string{"golangci-lint", "run"}

	if verbose.GetOr(false) {
		args = append(args, "--verbose")
	}

	if t, ok := timeout.Get(); ok {
		args = append(args, "--timeout", t)
	}

	ctr := goBase.Container().WithFile("/usr/local/bin/golangci-lint", golangciLint.File("/usr/bin/golangci-lint"))

	if src, ok := source.Get(); ok {
		const workdir = "/src"

		ctr = ctr.
			WithWorkdir(workdir).
			WithMountedDirectory(workdir, src)
	}

	ctr = ctr.WithExec(args)

	return ctr
}

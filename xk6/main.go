package main

import (
	"fmt"

	"github.com/containerd/containerd/platforms"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "grafana/xk6"

type Xk6 struct {
	// +private
	Ctr *Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	version Optional[string],

	// Custom image reference in "repository:tag" format to use as a base container.
	image Optional[string],

	// Custom container to use as a base container.
	container Optional[*Container],
) *Xk6 {
	var ctr *Container

	if v, ok := version.Get(); ok {
		ctr = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, v))
	} else if i, ok := image.Get(); ok {
		ctr = dag.Container().From(i)
	} else if c, ok := container.Get(); ok {
		ctr = c
	} else {
		ctr = dag.Container().From(defaultImageRepository)
	}

	return &Xk6{
		Ctr: ctr,
	}
}

func (m *Xk6) Container() *Container {
	return m.Ctr
}

// Set GOOS, GOARCH and GOARM environment variables.
func (m *Xk6) WithPlatform(platform Platform) *Xk6 {
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

	return &Xk6{ctr}
}

func (m *Xk6) Build(
	// k6 version to build (default: "latest")
	version Optional[string],

	// Extension to add to the k6 binary (format: <module[@version][=replacement]>)
	with Optional[[]string],

	// Add replacements to the go.mod file generated (format: <module=replacement>)
	replace Optional[[]string],

	// Target platform to build for (format: <os>/<arch>[/<variant>]) (e.g., "darwin/arm64/v7", "windows/amd64", "linux/arm64")
	platform Optional[Platform],
) *File {
	args := []string{"build", version.GetOr("latest")}

	for _, w := range with.GetOr([]string{}) {
		args = append(args, "--with", w)
	}

	for _, r := range replace.GetOr([]string{}) {
		args = append(args, "--replace", r)
	}

	if p, ok := platform.Get(); ok {
		m = m.WithPlatform(p)
	}

	return m.Ctr.WithExec(args).File("/xk6/k6")
}

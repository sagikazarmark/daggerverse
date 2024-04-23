// Build k6 with extensions.
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
	// +optional
	version string,

	// Custom image reference in "repository:tag" format to use as a base container.
	// +optional
	image string,

	// Custom container to use as a base container.
	// +optional
	container *Container,
) *Xk6 {
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

	return &Xk6{ctr}
}

func (m *Xk6) Container() *Container {
	return m.Ctr
}

// Set GOOS, GOARCH and GOARM environment variables.
func (m *Xk6) WithPlatform(
	// Target platform in "[os]/[platform]/[version]" format (e.g., "darwin/arm64/v7", "windows/amd64", "linux/arm64").
	platform Platform,
) *Xk6 {
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

// Build a custom k6 binary.
func (m *Xk6) Build(
	// k6 version to build (default: "latest")
	// +optional
	version string,

	// Extension to add to the k6 binary (format: <module[@version][=replacement]>)
	// +optional
	with []string,

	// Add replacements to the go.mod file generated (format: <module=replacement>)
	// +optional
	replace []string,

	// Target platform in "[os]/[platform]/[version]" format (e.g., "darwin/arm64/v7", "windows/amd64", "linux/arm64").
	// +optional
	platform Platform,
) *File {
	if version == "" {
		version = "latest"
	}

	args := []string{"build", version}

	for _, w := range with {
		args = append(args, "--with", w)
	}

	for _, r := range replace {
		args = append(args, "--replace", r)
	}

	if platform != "" {
		m = m.WithPlatform(platform)
	}

	return m.Ctr.WithExec(args).File("/xk6/k6")
}

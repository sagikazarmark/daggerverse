// Kubernetes native configuration management.
package main

import (
	"fmt"
	"path"
)

const (
	// defaultImageRepository is used when no image is specified.
	defaultImageRepository = "registry.k8s.io/kustomize/kustomize"

	// defaultVersion is used when no version is specified.
	//
	// (there is no latest tag published in the default image repository)
	defaultVersion = "v5.0.1"
)

type Kustomize struct {
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
) *Kustomize {
	var ctr *Container

	if version != "" {
		ctr = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	} else if image != "" {
		ctr = dag.Container().From(image)
	} else if container != nil {
		ctr = container
	} else {
		ctr = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, defaultVersion))
	}

	m := &Kustomize{ctr}

	return m
}

func (m *Kustomize) Container() *Container {
	return m.Ctr
}

// Build a kustomization target from a directory or URL.
func (m *Kustomize) Build(
	source *Directory,

	// Subdirectory within the source to use as the target.
	//
	// +optional
	dir string,
) *File {
	sourcePath := "/work/src"
	output := "/work/output.yaml"

	args := []string{"build", "--output", output}

	if dir != "" {
		args = append(args, path.Clean(dir))
	}

	return m.Ctr.
		WithWorkdir(sourcePath).
		WithMountedDirectory(sourcePath, source).
		WithExec(args).
		File(output)
}

// Edit a kustomization file.
func (m *Kustomize) Edit(source *Directory) *Edit {
	return &Edit{m.Ctr.WithWorkdir("/work").WithMountedDirectory("/work", source)}
}

// Edit a kustomization file.
type Edit struct {
	// +private
	Container *Container
}

func (m *Edit) Directory() *Directory {
	return m.Container.Directory("/work")
}

func (m *Edit) File() *File {
	return m.Container.File("/work/kustomization.yaml")
}

// Set the value of different fields in kustomization file.
func (m *Edit) Set() *Set {
	return &Set{m.Container}
}

// Set the value of different fields in kustomization file.
type Set struct {
	// +private
	Container *Container
}

// Sets one or more commonAnnotations in kustomization.yaml.
func (m *Set) Annotation(key string, value string) *Edit {
	return &Edit{m.Container.WithExec([]string{"edit", "set", "annotation", fmt.Sprintf("%s:%s", key, value)})}
}

// Set images and their new names, new tags or digests in the kustomization file.
func (m *Set) Image(image string) *Edit {
	return &Edit{m.Container.WithExec([]string{"edit", "set", "image", image})}
}

// Set the value of the namespace field in the kustomization file.
func (m *Set) Namespace(namespace string) *Edit {
	return &Edit{m.Container.WithExec([]string{"edit", "set", "namespace", namespace})}
}

// Set the value of the nameSuffix field in the kustomization file.
func (m *Set) Namesuffix(nameSuffix string) *Edit {
	return &Edit{m.Container.WithExec([]string{"edit", "set", "namesuffix", nameSuffix})}
}

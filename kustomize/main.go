// Kubernetes native configuration management.
package main

import (
	"dagger/kustomize/internal/dagger"
	"fmt"
	"path"
	"strings"
)

const (
	// defaultImageRepository is used when no image is specified.
	defaultImageRepository = "registry.k8s.io/kustomize/kustomize"

	// defaultVersion is used when no version is specified.
	//
	// (there is no latest tag published in the default image repository)
	defaultVersion = "v5.4.2"
)

type Kustomize struct {
	Container *dagger.Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	version string,

	// Custom container to use as a base container.
	//
	// +optional
	container *dagger.Container,
) *Kustomize {
	if container == nil {
		if version == "" {
			version = defaultVersion
		}

		container = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	}

	return &Kustomize{container}
}

func cleanPath(s string) string {
	s = path.Clean(s)

	for strings.HasPrefix(s, "../") {
		s = strings.TrimPrefix(s, "../")
	}

	return s
}

// Build a kustomization target from a directory or URL.
func (m *Kustomize) Build(
	source *dagger.Directory,

	// Subdirectory within the source to use as the target.
	//
	// +optional
	dir string,
) *dagger.File {
	sourcePath := "/work/src"
	output := "/work/output.yaml"

	args := []string{"kustomize", "build", "--output", output}

	if dir != "" {
		args = append(args, cleanPath(dir))
	}

	return m.Container.
		WithWorkdir(sourcePath).
		WithMountedDirectory(sourcePath, source).
		WithExec(args).
		File(output)
}

// Edit a kustomization file.
func (m *Kustomize) Edit(
	source *dagger.Directory,

	// Subdirectory within the source to use as the target.
	//
	// +optional
	dir string,
) *Edit {
	workdir := "/work"

	if dir != "" {
		workdir = path.Join(workdir, cleanPath(dir))
	}

	return &Edit{m.Container.WithMountedDirectory("/work", source).WithWorkdir(workdir)}
}

// Edit a kustomization file.
type Edit struct {
	// +private
	Container *dagger.Container
}

// Retrieve the source containing the modifications.
func (m *Edit) Directory() *dagger.Directory {
	return m.Container.Directory("/work")
}

// Set the value of different fields in kustomization file.
func (m *Edit) Set() *Set {
	return &Set{m.Container}
}

// Set the value of different fields in kustomization file.
type Set struct {
	// +private
	Container *dagger.Container
}

// Sets one or more commonAnnotations in kustomization.yaml.
func (m *Set) Annotation(key string, value string) *Edit {
	return &Edit{m.Container.WithExec([]string{"edit", "set", "annotation", fmt.Sprintf("%s:%s", key, value)})}
}

// Set images and their new names, new tags or digests in the kustomization file.
func (m *Set) Image(image string) *Edit {
	return &Edit{m.Container.WithExec([]string{"kustomize", "edit", "set", "image", image})}
}

// Set the value of the namespace field in the kustomization file.
func (m *Set) Namespace(namespace string) *Edit {
	return &Edit{m.Container.WithExec([]string{"kustomize", "edit", "set", "namespace", namespace})}
}

// Set the value of the nameSuffix field in the kustomization file.
func (m *Set) Namesuffix(nameSuffix string) *Edit {
	return &Edit{m.Container.WithExec([]string{"kustomize", "edit", "set", "namesuffix", nameSuffix})}
}

// Edit the value for an existing key in an existing Secret in the kustomization.yaml file.
func (m *Set) Secret(
	secret string,

	// Specify an existing key and a new value to update a Secret (i.e. mykey=newvalue).
	//
	// +optional
	fromLiteral []string,

	// Current namespace of the target Secret.
	//
	// +optional
	namespace string,

	// New namespace value for the target Secret.
	//
	// +optional
	newNamespace string,
) *Edit {
	args := []string{"kustomize", "edit", "set", "secret", secret}
	
	for _, literal := range fromLiteral {
		args = append(args, "--from-literal", literal)
	}
	
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}
	
	if newNamespace != "" {
		args = append(args, "--new-namespace", newNamespace)
	}

	return &Edit{m.Container.WithExec(args)}
}

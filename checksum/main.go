// Calculate and check the checksum of files.
package main

import (
	"fmt"
	"strings"
)

const alpineBaseImage = "alpine:latest"

// Calculate and check the checksum of files.
type Checksum struct{}

// Calculate the SHA-256 checksum of the given files.
func (m *Checksum) Sha256() *Sha256 {
	return &Sha256{}
}

type Sha256 struct{}

// Calculate the SHA-256 checksum of the given files.
func (m *Sha256) Calculate(
	// The files to calculate the checksum for.
	files []*File,
) *File {
	return calculate("sha256", files)
}

// Check the SHA-256 checksum of the given files.
func (m *Sha256) Check(
	// Checksum file.
	checksums *File,

	// The files to check the checksum if.
	files []*File,
) *Container {
	return check("sha256", checksums, files)
}

func calculate(algo string, files []*File) *File {
	dir := dag.Directory()

	for _, file := range files {
		dir = dir.WithFile("", file)
	}

	return calculateDirectory(algo, dir)
}

func calculateDirectory(algo string, dir *Directory) *File {
	const file = "/checksums.txt"

	cmd := []string{algo + "sum", "$(ls)", ">", file}

	return dag.Container().
		From(alpineBaseImage).
		WithWorkdir("/work").
		WithMountedDirectory("/work", dir).
		WithExec([]string{"sh", "-c", strings.Join(cmd, " ")}).
		File(file)
}

func check(algo string, checksums *File, files []*File) *Container {
	dir := dag.Directory()

	for _, file := range files {
		dir = dir.WithFile("", file)
	}

	return checkDirectory(algo, checksums, dir)
}

func checkDirectory(algo string, checksums *File, dir *Directory) *Container {
	dir = dir.WithFile("checksums.txt", checksums)

	return dag.Container().
		From(alpineBaseImage).
		WithWorkdir("/work").
		WithMountedDirectory("/work", dir).
		WithExec([]string{"sh", "-c", fmt.Sprintf("%ssum -w -c checksums.txt", algo)})
}

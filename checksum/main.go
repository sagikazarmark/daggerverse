// Calculate and check the checksum of files.
package main

import (
	"fmt"
	"path/filepath"
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

	// The name of the checksum file.
	// +optional
	// +default="checksums.txt"
	fileName string,
) *File {
	return calculate("sha256", fileName, files)
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

func calculate(algo string, fileName string, files []*File) *File {
	dir := dag.Directory()

	for _, file := range files {
		dir = dir.WithFile("", file)
	}

	return calculateDirectory(algo, fileName, dir)
}

func calculateDirectory(algo string, fileName string, dir *Directory) *File {
	if fileName == "" {
		fileName = "checksums.txt"
	}

	file := filepath.Join("/", filepath.Base(fileName))

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

package main

import (
	"context"
	"fmt"
)

const alpineBaseImage = "alpine:latest"

// Calculate and check the checksum of files.
type Checksum struct{}

// Calculate the SHA-256 checksum of the given files.
func (m *Checksum) SHA256(
	ctx context.Context,

	// The files to calculate the checksum for.
	files []*File,
) (string, error) {
	return m.calculate(ctx, "sha256", files)
}

func (m *Checksum) calculate(ctx context.Context, algo string, files []*File) (string, error) {
	dir := dag.Directory()

	for _, file := range files {
		dir = dir.WithFile("", file)
	}

	return m.calculateDirectory(ctx, algo, dir)
}

func (m *Checksum) calculateDirectory(ctx context.Context, algo string, dir *Directory) (string, error) {
	return dag.Container().
		From(alpineBaseImage).
		WithWorkdir("/work").
		WithMountedDirectory("/work", dir).
		WithExec([]string{"sh", "-c", fmt.Sprintf("%ssum $(ls)", algo)}).
		Stdout(ctx)
}

// Check the SHA-256 checksum of the given files.
func (m *Checksum) CheckSHA256(
	// Checksum content.
	checksums string,

	// The files to check the checksum if.
	files []*File,
) (*Container, error) {
	return m.check("sha256", checksums, files)
}

func (m *Checksum) check(algo string, checksums string, files []*File) (*Container, error) {
	dir := dag.Directory()

	for _, file := range files {
		dir = dir.WithFile("", file)
	}

	return m.checkDirectory(algo, checksums, dir)
}

func (m *Checksum) checkDirectory(algo string, checksums string, dir *Directory) (*Container, error) {
	dir = dir.WithNewFile("checksums.txt", checksums)

	return dag.Container().
		From(alpineBaseImage).
		WithWorkdir("/work").
		WithMountedDirectory("/work", dir).
		WithExec([]string{"sh", "-c", fmt.Sprintf("%ssum -w -c checksums.txt", algo)}), nil
}

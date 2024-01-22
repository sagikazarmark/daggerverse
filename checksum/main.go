package main

import (
	"context"
	"fmt"
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
	ctx context.Context,

	// The files to calculate the checksum for.
	files []*File,
) (string, error) {
	return calculate(ctx, "sha256", files)
}

// Check the SHA-256 checksum of the given files.
func (m *Sha256) Check(
	// Checksum content.
	checksums string,

	// The files to check the checksum if.
	files []*File,
) (*Container, error) {
	return check("sha256", checksums, files)
}

func calculate(ctx context.Context, algo string, files []*File) (string, error) {
	dir := dag.Directory()

	for _, file := range files {
		dir = dir.WithFile("", file)
	}

	return calculateDirectory(ctx, algo, dir)
}

func calculateDirectory(ctx context.Context, algo string, dir *Directory) (string, error) {
	return dag.Container().
		From(alpineBaseImage).
		WithWorkdir("/work").
		WithMountedDirectory("/work", dir).
		WithExec([]string{"sh", "-c", fmt.Sprintf("%ssum $(ls)", algo)}).
		Stdout(ctx)
}

func check(algo string, checksums string, files []*File) (*Container, error) {
	dir := dag.Directory()

	for _, file := range files {
		dir = dir.WithFile("", file)
	}

	return checkDirectory(algo, checksums, dir)
}

func checkDirectory(algo string, checksums string, dir *Directory) (*Container, error) {
	dir = dir.WithNewFile("checksums.txt", checksums)

	return dag.Container().
		From(alpineBaseImage).
		WithWorkdir("/work").
		WithMountedDirectory("/work", dir).
		WithExec([]string{"sh", "-c", fmt.Sprintf("%ssum -w -c checksums.txt", algo)}), nil
}

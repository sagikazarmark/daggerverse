package main

import (
	"context"
	"dagger/gh/internal/dagger"
	"errors"
	"fmt"
)

// Work with CodeQL.
func (m *Gh) Codeql(
	// +optional
	version string,
) *Codeql {
	return &Codeql{Gh: m}
}

type Codeql struct {
	// +private
	Version string

	// +private
	Gh *Gh
}

func (m *Codeql) container(token *dagger.Secret, repo string) *dagger.Container {
	version := m.Version
	if version == "" {
		version = "latest"
	}

	return m.Gh.container(token, repo).
		WithExec([]string{"gh", "extension", "install", "github/gh-codeql"}).
		WithExec([]string{"gh", "codeql", "set-version", version})
}

func (m *Codeql) Github() *CodeqlGithub {
	return &CodeqlGithub{Codeql: m}
}

type CodeqlGithub struct {
	// +private
	Codeql *Codeql
}

// Uploads a SARIF file to GitHub code scanning.
func (m *CodeqlGithub) UploadResults(
	ctx context.Context,

	// SARIF file to upload.
	sarif *dagger.File,

	// Name of the ref that was analyzed.
	//
	// +optional
	ref string,

	// SHA of commit that was analyzed.
	//
	// +optional
	commit string,

	// Disable waiting for GitHub to process the file.
	//
	// +optional
	noWaitForProcessing bool,

	// Timeout for waiting for GitHub to process the file.
	//
	// +optional
	waitForProcessingTimeout int,

	// GitHub token.
	//
	// +optional
	token *dagger.Secret,

	// GitHub repository (e.g. "owner/repo").
	//
	// +optional
	repo string,
) (string, error) {
	if repo == "" {
		repo = m.Codeql.Gh.Repository
	}

	if m.Codeql.Gh.Source == nil && (repo == "" || ref == "" || commit == "") {
		return "", errors.New("repo, ref and commit are required when no git repository source is available")
	}

	const sarifPath = "/work/codeql/report.sarif"

	args := []string{
		"gh", "codeql", "github", "upload-results",
		"--sarif", sarifPath,
	}

	if repo != "" {
		args = append(args, "--repository", repo)
	}

	if ref != "" {
		args = append(args, "--ref", ref)
	}

	if commit != "" {
		args = append(args, "--commit", commit)
	}

	if noWaitForProcessing {
		args = append(args, "--no-wait-for-processing")
	}

	if waitForProcessingTimeout > 0 {
		args = append(args, "--wait-for-processing-timeout", fmt.Sprint(waitForProcessingTimeout))
	}

	return m.Codeql.container(token, repo).
		WithMountedFile(sarifPath, sarif).
		WithExec(args).
		Stdout(ctx)
}

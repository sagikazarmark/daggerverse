// Verify provenance from SLSA compliant builders.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/google/go-github/v58/github"
	"github.com/hashicorp/go-getter"
)

type SlsaVerifier struct {
	// SLSA verifier version.
	// +private
	Version string
}

func New(
	// SLSA verifier version. (default: latest version)
	// +optional
	version string,
) *SlsaVerifier {
	return &SlsaVerifier{
		Version: version,
	}
}

// Verifies SLSA provenance on artifact blobs given as arguments (assuming same provenance).
func (m *SlsaVerifier) VerifyArtifact(
	ctx context.Context,

	// Artifacts to verify.
	artifacts []*File,

	// Provenance file.
	provenance *File,

	// Expected source repository that should have produced the binary, e.g. github.com/some/repo
	sourceURI string,

	// A workflow input provided by a user at trigger time in the format 'key=value'. (Only for 'workflow_dispatch' events on GitHub Actions).
	// +optional
	// buildWorkflowInput map[string]string,

	// The unique builder ID who created the provenance.
	// +optional
	builderID string,

	// Expected branch the binary was compiled from.
	// +optional
	sourceBranch string,

	// Expected tag the binary was compiled from.
	// +optional
	sourceTag string,

	// Expected version the binary was compiled from. Uses semantic version to match the tag.
	// +optional
	sourceVersionedTag string,
) (*Container, error) {
	if len(artifacts) == 0 {
		return nil, errors.New("no artifacts provided")
	}

	ctr, err := m.container(ctx)
	if err != nil {
		return nil, err
	}

	cmd := []string{
		"/usr/local/bin/slsa-verifier", "verify-artifact",
		"--print-provenance",
		"--provenance-path", "/work/provenance",
		"--source-uri", sourceURI,
	}

	if builderID != "" {
		cmd = append(cmd, "--builder-id", builderID)
	}

	if sourceBranch != "" {
		cmd = append(cmd, "--source-branch", sourceBranch)
	}

	if sourceTag != "" {
		cmd = append(cmd, "--source-tag", sourceTag)
	}

	if sourceVersionedTag != "" {
		cmd = append(cmd, "--source-versioned-tag", sourceVersionedTag)
	}

	artifactsDir := dag.Directory().With(func(d *Directory) *Directory {
		for _, artifact := range artifacts {
			d = d.WithFile("", artifact)
		}

		return d
	})

	cmd = append(cmd, "$(ls)")

	return ctr.
		WithWorkdir("/work/artifacts").
		WithMountedDirectory("/work/artifacts", artifactsDir).
		WithMountedFile("/work/provenance", provenance).
		WithExec([]string{"sh", "-c", strings.Join(cmd, " ")}, ContainerWithExecOpts{SkipEntrypoint: true}), nil
}

func (m *SlsaVerifier) container(ctx context.Context) (*Container, error) {
	binary, err := m.get(ctx, "")
	if err != nil {
		return nil, err
	}

	container := dag.Container().
		From("alpine:latest"). // TODO: make this configurable
		WithFile("/usr/local/bin/slsa-verifier", binary, ContainerWithFileOpts{Permissions: 0755}).
		WithEntrypoint([]string{"/usr/local/bin/slsa-verifier"})

	return container, nil
}

// Get the slsa-verifier binary.
// TODO: add checksum verification
func (m *SlsaVerifier) get(ctx context.Context, version string) (*File, error) {
	if version == "" {
		version = m.Version
	}

	// Get the latest version
	if version == "" {
		var err error

		version, err = getLatestVersion(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest version: %w", err)
		}
	}

	return downloadBinary(ctx, version)
}

// Get the latest slsa-verifier version.
func getLatestVersion(ctx context.Context) (string, error) {
	client := github.NewClient(nil)

	release, _, err := client.Repositories.GetLatestRelease(ctx, "slsa-framework", "slsa-verifier")
	if err != nil {
		return "", err
	}

	return *release.TagName, nil
}

const binaryDownloadURL = "https://github.com/slsa-framework/slsa-verifier/releases/download/v%s/slsa-verifier-%s-%s"

func downloadBinary(ctx context.Context, version string) (*File, error) {
	// slsa-verifier versions are prefixed with "v", but it looks better in parmeters without it.
	version = strings.TrimPrefix(version, "v")

	// TODO: support downloading binaries for other platforms.
	downloadURL := fmt.Sprintf(binaryDownloadURL, version, runtime.GOOS, runtime.GOARCH)
	// binaryName := fmt.Sprintf("slsa-verifier-%s-%s", runtime.GOOS, runtime.GOARCH)

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	client := getter.Client{
		Ctx:  ctx,
		Src:  downloadURL,
		Dst:  path.Join(pwd, "slsa-verifier"),
		Mode: getter.ClientModeFile,
	}

	err = client.Get()
	if err != nil {
		return nil, err
	}

	return dag.CurrentModule().WorkdirFile("slsa-verifier"), nil
}

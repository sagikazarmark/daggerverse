// Timoni is a package manager for Kubernetes, powered by CUE and inspired by Helm.
package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/google/go-github/v58/github"
	"github.com/hashicorp/go-getter"
)

type Timoni struct {
	// Timoni version.
	// +private
	Version string
}

func New(
	// Timoni version. (default: latest version)
	// +optional
	version string,
) *Timoni {
	return &Timoni{
		Version: version,
	}
}

func (m *Timoni) container(ctx context.Context) (*Container, error) {
	binary, err := m.get(ctx, "")
	if err != nil {
		return nil, err
	}

	container := dag.Container().
		From("alpine:latest"). // TODO: make this configurable
		WithFile("/usr/local/bin/timoni", binary, ContainerWithFileOpts{Permissions: 0755}).
		WithEntrypoint([]string{"/usr/local/bin/timoni"})

	return container, nil
}

// Get the timoni binary.
// TODO: add checksum verification
// TODO: add in-toto verification
func (m *Timoni) get(ctx context.Context, version string) (*File, error) {
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

// Get the latest timoni version.
func getLatestVersion(ctx context.Context) (string, error) {
	client := github.NewClient(nil)

	release, _, err := client.Repositories.GetLatestRelease(ctx, "stefanprodan", "timoni")
	if err != nil {
		return "", err
	}

	return *release.TagName, nil
}

const binaryDownloadURL = "https://github.com/stefanprodan/timoni/releases/download/v%s/timoni_%s_%s_%s.tar.gz"

func downloadBinary(ctx context.Context, version string) (*File, error) {
	// timoni versions are prefixed with "v", but it looks better in parmeters without it.
	version = strings.TrimPrefix(version, "v")

	// TODO: support downloading binaries for other platforms.
	downloadURL := fmt.Sprintf(binaryDownloadURL, version, version, runtime.GOOS, runtime.GOARCH)
	binaryName := fmt.Sprintf("timoni_%s_%s_%s", version, runtime.GOOS, runtime.GOARCH)

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	client := getter.Client{
		Ctx:  ctx,
		Src:  downloadURL,
		Dst:  pwd,
		Mode: getter.ClientModeDir,
	}

	err = client.Get()
	if err != nil {
		return nil, err
	}

	return dag.CurrentModule().WorkdirFile(path.Join(binaryName, "timoni")), nil
}

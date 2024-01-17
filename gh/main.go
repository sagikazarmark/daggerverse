package main

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/google/go-github/v58/github"
	"github.com/hashicorp/go-getter"
)

type Gh struct {
	// GitHub CLI version.
	// +private
	Version string

	// GitHub token.
	// +private
	Token *Secret

	// GitHub repository (e.g. "owner/Repo").
	// +private
	Repo string
}

func New(
	// GitHub CLI version. (default: latest version)
	// +optional
	version string,

	// GitHub token.
	// +optional
	token *Secret,

	// GitHub repository (e.g. "owner/repo").
	// +optional
	repo string,
) *Gh {
	return &Gh{
		Version: version,
		Token:   token,
		Repo:    repo,
	}
}

// Run a GitHub CLI command (accepts a single command string without "gh").
func (m *Gh) Run(
	ctx context.Context,

	// Command to run.
	cmd string,

	// GitHub token.
	// +optional
	token *Secret,

	// GitHub repository (e.g. "owner/repo").
	// +optional
	repo string,
) (*Container, error) {
	ctr, err := m.container(ctx, token, repo)
	if err != nil {
		return nil, nil
	}

	return ctr.WithExec([]string{"sh", "-c", strings.Join([]string{"/usr/local/bin/gh", cmd}, " ")}, ContainerWithExecOpts{SkipEntrypoint: true}), nil
}

// Run a GitHub CLI command (accepts a list of arguments without "gh").
func (m *Gh) Exec(
	ctx context.Context,

	// Arguments to pass to GitHub CLI.
	args []string,

	// GitHub token.
	// +optional
	token *Secret,

	// GitHub repository (e.g. "owner/repo").
	// +optional
	repo string,
) (*Container, error) {
	ctr, err := m.container(ctx, token, repo)
	if err != nil {
		return nil, nil
	}

	return ctr.WithExec(args), nil
}

func (m *Gh) container(ctx context.Context, token *Secret, repo string) (*Container, error) {
	if token == nil {
		token = m.Token
	}

	if repo == "" {
		repo = m.Repo
	}

	binary, err := m.get(ctx, "", token)
	if err != nil {
		return nil, err
	}

	container := dag.Container().
		From("alpine/git:latest"). // TODO: make this configurable
		WithEnvVariable("GH_PROMPT_DISABLED", "true").
		WithEnvVariable("GH_NO_UPDATE_NOTIFIER", "true").
		WithFile("/usr/local/bin/gh", binary).
		With(func(c *Container) *Container {
			if token != nil {
				c = c.WithSecretVariable("GITHUB_TOKEN", token)
			}

			if repo != "" {
				c = c.WithEnvVariable("GH_REPO", repo)
			}

			return c
		}).
		WithEntrypoint([]string{"/usr/local/bin/gh"})

	return container, nil
}

// Get the GitHub CLI binary.
func (m *Gh) Get(
	ctx context.Context,

	// GitHub CLI version. (default: latest version)
	// +optional
	version string,

	// GitHub token. (May be used to get the latest version from the GitHub API)
	// +optional
	token *Secret,
) (*File, error) {
	return m.get(ctx, version, token)
}

// Get the GitHub CLI binary.
func (m *Gh) get(ctx context.Context, version string, token *Secret) (*File, error) {
	if version == "" {
		version = m.Version
	}

	// Get the latest version
	if version == "" {
		if token == nil {
			token = m.Token
		}

		var err error

		version, err = getLatestVersion(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest GitHub CLI version: %w", err)
		}
	}

	return downloadBinary(ctx, version)
}

// Get the latest GitHub CLI version.
func getLatestVersion(ctx context.Context, token *Secret) (string, error) {
	client := github.NewClient(nil)

	if token != nil {
		tokenContent, err := token.Plaintext(ctx)
		if err != nil {
			return "", err
		}

		client = client.WithAuthToken(tokenContent)
	}

	release, _, err := client.Repositories.GetLatestRelease(ctx, "cli", "cli")
	if err != nil {
		return "", err
	}

	return *release.TagName, nil
}

const binaryDownloadURL = "https://github.com/cli/cli/releases/download/v%s/gh_%s_%s_%s.tar.gz"

func downloadBinary(ctx context.Context, version string) (*File, error) {
	// GitHub CLI versions are prefixed with "v", but it looks better in parmeters without it.
	version = strings.TrimPrefix(version, "v")

	// TODO: support downloading binaries for other platforms.
	downloadURL := fmt.Sprintf(binaryDownloadURL, version, version, runtime.GOOS, runtime.GOARCH)
	binaryName := fmt.Sprintf("gh_%s_%s_%s", version, runtime.GOOS, runtime.GOARCH)

	client := getter.Client{
		Ctx:  ctx,
		Src:  downloadURL,
		Dst:  "/tmp",
		Mode: getter.ClientModeDir,
	}

	err := client.Get()
	if err != nil {
		return nil, err
	}

	return dag.Host().File(path.Join("/tmp", binaryName, "bin/gh")), nil
}

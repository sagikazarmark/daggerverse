// GitHub CLI
package main

import (
	"errors"
	"strings"
	"time"
)

type Gh struct {
	// GitHub token.
	//
	// +private
	Token *Secret

	// GitHub repository (e.g. "owner/repo").
	//
	// +private
	Repository string

	// Git repository source (with .git directory).
	Source *Directory
}

func New(
	// GitHub token.
	//
	// +optional
	token *Secret,

	// GitHub repository (e.g. "owner/repo").
	//
	// +optional
	repo string,

	// Git repository source (with .git directory).
	//
	// +optional
	source *Directory,
) (*Gh, error) {
	return &Gh{
		Token:      token,
		Repository: repo,
	}, nil
}

// Set a GitHub token.
func (m *Gh) WithToken(
	// GitHub token.
	token *Secret,
) *Gh {
	gh := *m

	gh.Token = token

	return &gh
}

// Set a GitHub repository as context.
func (m *Gh) WithRepo(
	// GitHub repository (e.g. "owner/repo").
	repo string,
) (*Gh, error) {
	gh := *m

	gh.Repository = repo

	return &gh, nil
}

// Load a Git repository source (with .git directory).
func (m *Gh) WithSource(
	// Git repository source (with .git directory).
	source *Directory,
) *Gh {
	gh := *m

	gh.Source = source

	return &gh
}

// Clone a GitHub repository.
func (m *Gh) Clone(
	// GitHub repository (e.g. "owner/repo").
	//
	// +optional
	repo string,
) (*Gh, error) {
	if repo == "" {
		repo = m.Repository
	}

	if repo == "" {
		return nil, errors.New("no repository specified")
	}

	return m.WithSource(m.Repo().Clone(repo, nil, nil)), nil
}

// Run a GitHub CLI command (accepts a single command string without "gh").
func (m *Gh) Run(
	// Command to run.
	cmd string,

	// GitHub token.
	//
	// +optional
	token *Secret,

	// GitHub repository (e.g. "owner/repo").
	//
	// +optional
	repo string,
) *Container {
	return m.container(token, repo).WithExec([]string{"sh", "-c", strings.Join([]string{"gh", cmd}, " ")})
}

// Run a GitHub CLI command (accepts a list of arguments without "gh").
func (m *Gh) Exec(
	// Arguments to pass to GitHub CLI.
	args []string,

	// GitHub token.
	//
	// +optional
	token *Secret,

	// GitHub repository (e.g. "owner/repo").
	//
	// +optional
	repo string,
) *Container {
	return m.container(token, repo).WithExec(args)
}

// Open an interactive terminal.
func (m *Gh) Terminal(
	// GitHub token.
	//
	// +optional
	token *Secret,

	// GitHub repository (e.g. "owner/repo").
	//
	// +optional
	repo string,
) *Terminal {
	return m.container(token, repo).Terminal()
}

func (m *Gh) base() *Container {
	return dag.
		Wolfi().
		Container(WolfiContainerOpts{
			Packages: []string{
				"gh",
				"git",
			},
		}).
		WithEnvVariable("GH_PROMPT_DISABLED", "true").
		WithEnvVariable("GH_NO_UPDATE_NOTIFIER", "true").
		WithExec([]string{"gh", "auth", "setup-git", "--force", "--hostname", "github.com"}) // Use force to avoid network call and cache setup even when no token is provided.
}

func (m *Gh) container(token *Secret, repo string) *Container {
	if token == nil {
		token = m.Token
	}

	if repo == "" {
		repo = m.Repository
	}

	return m.base().
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		With(func(c *Container) *Container {
			if token != nil {
				c = c.WithSecretVariable("GITHUB_TOKEN", token)
			}

			if repo != "" {
				c = c.WithEnvVariable("GH_REPO", repo)
			}

			if m.Source != nil {
				c = c.
					WithWorkdir("/work/repo").
					WithMountedDirectory("/work/repo", m.Source)
			}

			return c
		})
}

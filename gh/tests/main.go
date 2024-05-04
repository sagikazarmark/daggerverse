package main

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct {
	GitHubToken *Secret
}

func New(
	// +optional
	githubToken *Secret,
) *Tests {
	return &Tests{
		GitHubToken: githubToken,
	}
}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Help)
	p.Go(m.Clone)

	return p.Wait()
}

func (m *Tests) Help(ctx context.Context) error {
	_, err := dag.Gh().Run("--help").Sync(ctx)

	return err
}

func (m *Tests) Clone(ctx context.Context) error {
	_, err := dag.Gh().
		With(func(g *Gh) *Gh {
			if m.GitHubToken != nil {
				g = g.WithToken(m.GitHubToken)
			}

			return g
		}).
		WithRepo("sagikazarmark/daggerverse").
		Clone().
		Source().
		Sync(ctx)

	return err
}

func (m *Tests) WithGitExec(ctx context.Context) error {
	_, err := dag.Gh().
		With(func(g *Gh) *Gh {
			if m.GitHubToken != nil {
				g = g.WithToken(m.GitHubToken)
			}

			return g
		}).
		WithRepo("sagikazarmark/daggerverse").
		Clone().
		WithGitExec([]string{"checkout", "-b", "gh-test-checkout"}).
		Source().
		Sync(ctx)

	return err
}

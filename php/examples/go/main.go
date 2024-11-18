package main

import (
	"context"
	_ "dagger/helm/examples/php/internal/dagger"

	"github.com/sourcegraph/conc/pool"
)

type Examples struct{}

// All executes all examples.
func (m *Examples) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Php_Composer)
	p.Go(m.Php_Extension)

	return p.Wait()
}

func (m *Examples) Php_Composer(ctx context.Context) error {
	_, err := dag.Php().
		WithComposer(). // This is optional: will be installed automatically
		WithComposerPackage("phpstan/phpstan").
		Container().
		WithExec([]string{"phpstan", "analyse", "--help"}).
		Sync(ctx)

	return err
}

func (m *Examples) Php_Extension(ctx context.Context) error {
	_, err := dag.Php().
		WithExtensionInstaller(). // This is optional: will be installed automatically
		WithExtension("zip").
		Container().
		WithExec([]string{"php", "-m"}).
		Sync(ctx)

	return err
}

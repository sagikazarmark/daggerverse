package main

import (
	"context"
	"dagger/xcaddy/tests/internal/dagger"
	"fmt"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Test)
	p.Go(m.Version)
	p.Go(m.WithVersion)
	p.Go(m.WithModule().All)

	return p.Wait()
}

func (m *Tests) Test(ctx context.Context) error {
	binary := dag.Xcaddy().Build().Binary()

	caddyService := dag.Container().
		From("caddy").
		WithFile("/usr/bin/caddy", binary).
		WithFile("/etc/caddy/Caddyfile", dag.CurrentModule().Source().File("Caddyfile")).
		WithExposedPort(80).
		AsService()

	actual, err := dag.Container().
		From("alpine").
		WithExec([]string{"apk", "add", "curl"}).
		WithServiceBinding("caddy", caddyService).
		WithExec([]string{"curl", "http://caddy"}).
		Stdout(ctx)
	if err != nil {
		return err
	}

	if actual != "Hello, world!" {
		return fmt.Errorf("unexpected response from caddy: %q", actual)
	}

	return nil
}

func (m *Tests) Version(ctx context.Context) error {
	binary := dag.Xcaddy().Build(XcaddyBuildOpts{Version: "latest"}).Binary()

	_, err := dag.Container().
		From("alpine").
		WithFile("/usr/local/bin/caddy", binary).
		WithExec([]string{"caddy", "version"}).
		Sync(ctx)

	return err
}

func (m *Tests) WithVersion(ctx context.Context) error {
	binary := dag.Xcaddy().Build(XcaddyBuildOpts{Version: "latest"}).WithVersion("v2.8.4").Binary()

	_, err := dag.Container().
		From("alpine").
		WithFile("/usr/local/bin/caddy", binary).
		// TODO: make sure version is correct
		WithExec([]string{"caddy", "version"}).
		Sync(ctx)

	return err
}

func (m *Tests) WithModule() *WithModule {
	return &WithModule{}
}

type WithModule struct{}

// All executes all tests.
func (m *WithModule) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Module)
	p.Go(m.Version)
	p.Go(m.Replacement)

	return p.Wait()
}

func (m *WithModule) Module(ctx context.Context) error {
	binary := dag.Xcaddy().
		Build().
		WithModule("github.com/sagikazarmark/caddy-fs-s3").
		Binary()

	_, err := dag.Container().
		From("alpine").
		WithFile("/usr/local/bin/caddy", binary).
		WithExec([]string{"caddy", "version"}).
		Sync(ctx)

	return err
}

func (m *WithModule) Version(ctx context.Context) error {
	binary := dag.Xcaddy().
		Build().
		WithModule("github.com/sagikazarmark/caddy-fs-s3", dagger.XcaddyBuildWithModuleOpts{
			Version: "v0.3.1",
		}).
		Binary()

	_, err := dag.Container().
		From("alpine").
		WithFile("/usr/local/bin/caddy", binary).
		WithExec([]string{"caddy", "version"}).
		Sync(ctx)

	return err
}

func (m *WithModule) Replacement(ctx context.Context) error {
	binary := dag.Xcaddy().
		Build().
		WithModule("github.com/sagikazarmark/caddy-fs-s3", dagger.XcaddyBuildWithModuleOpts{
			// TODO: lock to specific version?
			Replacement: dag.Git("https://github.com/sagikazarmark/caddy-fs-s3.git").Branch("main").Tree(),
		}).
		Binary()

	_, err := dag.Container().
		From("alpine").
		WithFile("/usr/local/bin/caddy", binary).
		WithExec([]string{"caddy", "version"}).
		Sync(ctx)

	return err
}

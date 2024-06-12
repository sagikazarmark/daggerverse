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
	p.Go(m.Plugin().All)

	return p.Wait()
}

func (m *Tests) Test(ctx context.Context) error {
	caddyService := dag.Xcaddy().Build().Container().
		From("caddy").
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
	binary := dag.Xcaddy().Build(XcaddyBuildOpts{Version: "v2.8.4"}).Binary()

	_, err := dag.Container().
		From("alpine").
		WithFile("/usr/local/bin/caddy", binary).
		WithExec([]string{"caddy", "version"}).
		Sync(ctx)

	return err
}

func (m *Tests) Plugin() *Plugin {
	return &Plugin{}
}

type Plugin struct{}

// All executes all tests.
func (m *Plugin) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Module)
	p.Go(m.Version)
	p.Go(m.Replacement)

	return p.Wait()
}

func (m *Plugin) Module(ctx context.Context) error {
	binary := dag.Xcaddy().
		Build().
		Plugin("github.com/sagikazarmark/caddy-fs-s3").
		Binary()

	_, err := dag.Container().
		From("alpine").
		WithFile("/usr/local/bin/caddy", binary).
		WithExec([]string{"caddy", "version"}).
		Sync(ctx)

	return err
}

func (m *Plugin) Version(ctx context.Context) error {
	binary := dag.Xcaddy().
		Build().
		Plugin("github.com/sagikazarmark/caddy-fs-s3", dagger.XcaddyBuildPluginOpts{
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

func (m *Plugin) Replacement(ctx context.Context) error {
	binary := dag.Xcaddy().
		Build().
		Plugin("github.com/sagikazarmark/caddy-fs-s3", dagger.XcaddyBuildPluginOpts{
			// TODO: lock to specific version?
			Replacement: dag.Git("https://github.com/sagikazarmark/caddy-fs-s3.git").Tag("v0.4.0").Tree(),
		}).
		Binary()

	_, err := dag.Container().
		From("alpine").
		WithFile("/usr/local/bin/caddy", binary).
		WithExec([]string{"caddy", "version"}).
		Sync(ctx)

	return err
}

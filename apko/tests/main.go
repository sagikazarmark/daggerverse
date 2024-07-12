package main

import (
	"context"
	"dagger/apko/tests/internal/dagger"
	"fmt"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Build)
	p.Go(m.Publish)
	p.Go(m.Wolfi)

	return p.Wait()
}

func (m *Tests) Build(ctx context.Context) error {
	result := dag.Apko().Build(dag.CurrentModule().Source().File("testdata/wolfi-base.yaml"), "latest")

	_, err := dag.Container().Import(result.File()).WithExec([]string{"cat", "/etc/apk/repositories"}).Sync(ctx)

	return err
}

func (m *Tests) Publish(ctx context.Context) error {
	registry, ca, err := registryService(ctx)
	if err != nil {
		return err
	}

	password := dag.SetSecret("registry-password", "password")

	return dag.Apko(dagger.ApkoOpts{
		Container: dag.Container().
			From("cgr.dev/chainguard/apko").
			WithMountedFile("/etc/ssl/certs/test.pem", ca).
			WithServiceBinding("zot", registry),
	}).
		WithRegistryAuth("zot:8080", "username", password).
		Publish(ctx, dag.CurrentModule().Source().File("testdata/wolfi-base.yaml"), "zot:8080/wolfi-base")
}

func (m *Tests) Wolfi(ctx context.Context) error {
	const expected = "https://packages.wolfi.dev/os\n"

	actual, err := dag.Apko().Wolfi().Container().WithExec([]string{"cat", "/etc/apk/repositories"}).Stdout(ctx)
	if err != nil {
		return err
	}

	if actual != expected {
		return fmt.Errorf("expected %q, got %q", expected, actual)
	}

	return nil
}

func (m *Tests) Alpine(ctx context.Context) error {
	const expected = "https://dl-cdn.alpinelinux.org/alpine/edge/main\n"

	actual, err := dag.Apko().Alpine().Container().WithExec([]string{"cat", "/etc/apk/repositories"}).Stdout(ctx)
	if err != nil {
		return err
	}

	if actual != expected {
		return fmt.Errorf("expected %q, got %q", expected, actual)
	}

	return nil
}

func registryService(ctx context.Context) (*dagger.Service, *dagger.File, error) {
	const zotRepositoryTemplate = "ghcr.io/project-zot/zot-%s-%s"
	const zotVersion = "v2.1.0"

	mkcert := dag.Container().
		From("cgr.dev/chainguard/wolfi-base").
		WithExec([]string{"apk", "add", "mkcert"}).
		WithExec([]string{"mkcert", "-install"}).
		WithWorkdir("/work").
		WithExec([]string{"mkcert", "zot"})

	platform, err := dag.DefaultPlatform(ctx)
	if err != nil {
		return nil, nil, err
	}

	platformArgs := strings.Split(string(platform), "/")

	zotRepository := fmt.Sprintf(zotRepositoryTemplate, platformArgs[0], platformArgs[1])

	return dag.Container().
		From(fmt.Sprintf("%s:%s", zotRepository, zotVersion)).
		WithExposedPort(8080).
		WithMountedDirectory("/etc/zot", dag.CurrentModule().Source().Directory("./testdata/zot")).
		WithMountedDirectory("/etc/zot/tls", mkcert.Directory("/work")).
		WithExec([]string{"serve", "/etc/zot/config.json"}, dagger.ContainerWithExecOpts{
			UseEntrypoint: true,
		}).
		AsService(), mkcert.File("/root/.local/share/mkcert/rootCA.pem"), nil
}

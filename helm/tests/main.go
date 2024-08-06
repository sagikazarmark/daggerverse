package main

import (
	"context"
	"dagger/helm/tests/internal/dagger"
	"fmt"
	"slices"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

const helmVersion = "3.13.2"

func newHelm() *dagger.Helm {
	return dag.Helm(dagger.HelmOpts{
		Version: helmVersion,
	})
}

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Create)
	p.Go(m.Lint)
	p.Go(m.Login)
	p.Go(m.Package)
	p.Go(m.Push)
	p.Go(m.ChartLint)
	p.Go(m.ChartPackage)
	p.Go(m.ChartPublish)

	return p.Wait()
}

func (m *Tests) Create(ctx context.Context) error {
	dir := newHelm().Create("foo").Directory()

	entries, err := dir.Entries(ctx)
	if err != nil {
		return err
	}

	if !slices.Contains(entries, "Chart.yaml") {
		return fmt.Errorf("expected chart directory to contain Chart.yaml")
	}

	return nil
}

// TODO: improve this test
func (m *Tests) Lint(ctx context.Context) error {
	_, err := newHelm().
		Lint(dag.CurrentModule().Source().Directory("./testdata/charts/package")).
		Sync(ctx)

	return err
}

// TODO: improve this test
func (m *Tests) Login(ctx context.Context) error {
	registry, err := registryService(ctx)
	if err != nil {
		return err
	}

	password := dag.SetSecret("registry-password", "password")

	_, err = dag.Helm(dagger.HelmOpts{
		Container: newHelm().Container().
			WithServiceBinding("zot", registry),
	}).
		Login("zot:8080", "username", password, dagger.HelmLoginOpts{
			Insecure: true,
		}).Container().Sync(ctx)

	return err
}

// TODO: improve this test
func (m *Tests) Package(ctx context.Context) error {
	_, err := newHelm().
		Package(dag.CurrentModule().Source().Directory("./testdata/charts/package")).
		Sync(ctx)

	return err
}

// TODO: improve this test
func (m *Tests) Push(ctx context.Context) error {
	registry, err := registryService(ctx)
	if err != nil {
		return err
	}

	pkg := newHelm().Package(dag.CurrentModule().Source().Directory("./testdata/charts/package"))

	password := dag.SetSecret("registry-password", "password")

	return dag.Helm(dagger.HelmOpts{
		Container: newHelm().Container().WithServiceBinding("zot", registry),
	}).
		WithRegistryAuth("zot:8080", "username", password).
		Push(ctx, pkg, "oci://zot:8080/helm-charts", dagger.HelmPushOpts{
			PlainHTTP: true,
		})
}

// TODO: improve this test
func (m *Tests) ChartLint(ctx context.Context) error {
	_, err := newHelm().
		Chart(dag.CurrentModule().Source().Directory("./testdata/charts/package")).
		Lint().
		Sync(ctx)

	return err
}

// TODO: improve this test
func (m *Tests) ChartPackage(ctx context.Context) error {
	_, err := newHelm().
		Chart(dag.CurrentModule().Source().Directory("./testdata/charts/package")).
		Package().
		File().
		Sync(ctx)

	return err
}

// TODO: improve this test
func (m *Tests) ChartPublish(ctx context.Context) error {
	registry, err := registryService(ctx)
	if err != nil {
		return err
	}

	password := dag.SetSecret("registry-password", "password")

	return dag.Helm(dagger.HelmOpts{
		Container: newHelm().Container().WithServiceBinding("zot", registry),
	}).
		Chart(dag.CurrentModule().Source().Directory("./testdata/charts/package")).
		Package().
		WithRegistryAuth("zot:8080", "username", password).
		Publish(ctx, "oci://zot:8080/helm-charts", dagger.HelmPackagePublishOpts{
			PlainHTTP: true,
		})
}

func registryService(ctx context.Context) (*dagger.Service, error) {
	const zotRepositoryTemplate = "ghcr.io/project-zot/zot-%s-%s"
	const zotVersion = "v2.0.0"

	platform, err := dag.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	platformArgs := strings.Split(string(platform), "/")

	zotRepository := fmt.Sprintf(zotRepositoryTemplate, platformArgs[0], platformArgs[1])

	return dag.Container().
		From(fmt.Sprintf("%s:%s", zotRepository, zotVersion)).
		WithExposedPort(8080).
		WithMountedDirectory("/etc/zot", dag.CurrentModule().Source().Directory("./testdata/zot")).
		WithExec([]string{"serve", "/etc/zot/config.json"}, dagger.ContainerWithExecOpts{UseEntrypoint: true}).
		AsService(), nil
}

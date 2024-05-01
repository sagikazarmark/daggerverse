package main

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

const helmVersion = "3.13.2"

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Create)
	p.Go(m.Lint)
	p.Go(m.LintOld)
	p.Go(m.Package)
	p.Go(m.PackageOld)
	// p.Go(m.LoginAndPush)
	// p.Go(m.LoginAndPushOld)
	p.Go(m.LoginOld)

	return p.Wait()
}

func (m *Tests) Create(ctx context.Context) error {
	dir := dag.Helm(HelmOpts{
		Version: helmVersion,
	}).
		Create("foo")

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
	_, err := dag.Helm(HelmOpts{
		Version: helmVersion,
	}).
		Chart(dag.CurrentModule().Source().Directory("./testdata/charts/package")).
		Lint().
		Sync(ctx)

	return err
}

// TODO: improve this test
// TODO: improve naming
func (m *Tests) LintOld(ctx context.Context) error {
	_, err := dag.Helm(HelmOpts{
		Version: helmVersion,
	}).
		Lint(dag.CurrentModule().Source().Directory("./testdata/charts/package")).
		Sync(ctx)

	return err
}

// TODO: improve this test
func (m *Tests) Package(ctx context.Context) error {
	_, err := dag.Helm(HelmOpts{
		Version: helmVersion,
	}).
		Chart(dag.CurrentModule().Source().Directory("./testdata/charts/package")).
		Package().
		File().
		Sync(ctx)

	return err
}

// TODO: improve this test
func (m *Tests) PackageOld(ctx context.Context) error {
	_, err := dag.Helm(HelmOpts{
		Version: helmVersion,
	}).
		Package(dag.CurrentModule().Source().Directory("./testdata/charts/package")).
		Sync(ctx)

	return err
}

// TODO: improve this test
func (m *Tests) LoginAndPush(ctx context.Context) error {
	const zotRepositoryTemplate = "ghcr.io/project-zot/zot-%s-%s"
	const zotVersion = "v2.0.0"

	platform, err := dag.DefaultPlatform(ctx)
	if err != nil {
		return err
	}

	platformArgs := strings.Split(string(platform), "/")

	zotRepository := fmt.Sprintf(zotRepositoryTemplate, platformArgs[0], platformArgs[1])

	registry := dag.Container().
		From(fmt.Sprintf("%s:%s", zotRepository, zotVersion)).
		WithExposedPort(8080).
		WithMountedDirectory("/etc/zot", dag.CurrentModule().Source().Directory("./testdata/zot")).
		WithExec([]string{"serve", "/etc/zot/config.json"}).
		AsService()

	password := dag.SetSecret("registry-password", "password")

	_, err = dag.Helm(HelmOpts{
		Container: dag.Helm(HelmOpts{
			Version: helmVersion,
		}).Container().WithServiceBinding("zot", registry),
	}).
		Chart(dag.CurrentModule().Source().Directory("./testdata/charts/package")).
		Package().
		WithRegistryAuth("zot:8080", "username", password, HelmPackageWithRegistryAuthOpts{
			Insecure: true,
		}).
		Publish(ctx, "oci://zot:8080/helm-charts", HelmPackagePublishOpts{
			PlainHTTP: true,
		})

	return err
}

// TODO: improve this test
func (m *Tests) LoginAndPushOld(ctx context.Context) error {
	const zotRepositoryTemplate = "ghcr.io/project-zot/zot-%s-%s"
	const zotVersion = "v2.0.0"

	platform, err := dag.DefaultPlatform(ctx)
	if err != nil {
		return err
	}

	platformArgs := strings.Split(string(platform), "/")

	zotRepository := fmt.Sprintf(zotRepositoryTemplate, platformArgs[0], platformArgs[1])

	helm := dag.Helm(HelmOpts{
		Version: helmVersion,
	})

	pkg := helm.Package(dag.CurrentModule().Source().Directory("./testdata/charts/package"))

	registry := dag.Container().
		From(fmt.Sprintf("%s:%s", zotRepository, zotVersion)).
		WithExposedPort(8080).
		WithMountedDirectory("/etc/zot", dag.CurrentModule().Source().Directory("./testdata/zot")).
		WithExec([]string{"serve", "/etc/zot/config.json"}).
		AsService()

	password := dag.SetSecret("registry-password", "password")

	_, err = dag.Helm(HelmOpts{
		Container: helm.Container().
			WithServiceBinding("zot2", registry),
	}).
		Login("zot2:8080", "username", password, HelmLoginOpts{
			Insecure: true,
		}).
		Push(ctx, pkg, "oci://zot2:8080/helm-charts", HelmPushOpts{
			PlainHTTP: true,
		})

	return err
}

// TODO: improve this test
func (m *Tests) LoginOld(ctx context.Context) error {
	const zotRepositoryTemplate = "ghcr.io/project-zot/zot-%s-%s"
	const zotVersion = "v2.0.0"

	platform, err := dag.DefaultPlatform(ctx)
	if err != nil {
		return err
	}

	platformArgs := strings.Split(string(platform), "/")

	zotRepository := fmt.Sprintf(zotRepositoryTemplate, platformArgs[0], platformArgs[1])

	helm := dag.Helm(HelmOpts{
		Version: helmVersion,
	})

	registry := dag.Container().
		From(fmt.Sprintf("%s:%s", zotRepository, zotVersion)).
		WithExposedPort(8080).
		WithMountedDirectory("/etc/zot", dag.CurrentModule().Source().Directory("./testdata/zot")).
		WithExec([]string{"serve", "/etc/zot/config.json"}).
		AsService()

	password := dag.SetSecret("registry-password", "password")

	_, err = dag.Helm(HelmOpts{
		Container: helm.Container().
			WithServiceBinding("zot", registry),
	}).
		Login("zot:8080", "username", password, HelmLoginOpts{
			Insecure: true,
		}).Container().Sync(ctx)

	return err
}

package main

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
)

func (m *Ci) Helm(ctx context.Context) error {
	var group errgroup.Group

	const helmVersion = "3.13.2"

	// Lint
	// TODO: improve this test
	group.Go(func() error {
		_, err := dag.Helm(HelmOpts{
			Version: helmVersion,
		}).
			Chart(dag.CurrentModule().Source().Directory("./testdata/helm/charts/package")).
			Lint().
			Sync(ctx)

		return err
	})

	// Lint (old)
	// TODO: improve this test
	group.Go(func() error {
		_, err := dag.Helm(HelmOpts{
			Version: helmVersion,
		}).
			Lint(dag.CurrentModule().Source().Directory("./testdata/helm/charts/package")).
			Sync(ctx)

		return err
	})

	// Package
	// TODO: improve this test
	group.Go(func() error {
		_, err := dag.Helm(HelmOpts{
			Version: helmVersion,
		}).
			Chart(dag.CurrentModule().Source().Directory("./testdata/helm/charts/package")).
			Package().
			File().
			Sync(ctx)

		return err
	})

	// Package (old)
	// TODO: improve this test
	group.Go(func() error {
		_, err := dag.Helm(HelmOpts{
			Version: helmVersion,
		}).
			Package(dag.CurrentModule().Source().Directory("./testdata/helm/charts/package")).
			Sync(ctx)

		return err
	})

	// Login & push
	// TODO: improve this test
	group.Go(func() error {
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
			WithMountedDirectory("/etc/zot", dag.CurrentModule().Source().Directory("./testdata/helm/zot")).
			WithExec([]string{"serve", "/etc/zot/config.json"}).
			AsService()

		password := dag.SetSecret("registry-password", "password")

		_, err = dag.Helm(HelmOpts{
			Container: dag.Helm(HelmOpts{
				Version: helmVersion,
			}).Container().WithServiceBinding("zot", registry),
		}).
			Chart(dag.CurrentModule().Source().Directory("./testdata/helm/charts/package")).
			Package().
			WithRegistryAuth("zot:8080", "username", password, HelmPackageWithRegistryAuthOpts{
				Insecure: true,
			}).
			Publish(ctx, "oci://zot:8080/helm-charts", HelmPackagePublishOpts{
				PlainHTTP: true,
			})

		return err
	})

	// Login & push (old)
	// TODO: improve this test
	// group.Go(func() error {
	// 	const zotRepositoryTemplate = "ghcr.io/project-zot/zot-%s-%s"
	// 	const zotVersion = "v2.0.0"
	//
	// 	platform, err := dag.DefaultPlatform(ctx)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	platformArgs := strings.Split(string(platform), "/")
	//
	// 	zotRepository := fmt.Sprintf(zotRepositoryTemplate, platformArgs[0], platformArgs[1])
	//
	// 	helm := dag.Helm(HelmOpts{
	// 		Version: helmVersion,
	// 	})
	//
	// 	pkg := helm.Package(dag.CurrentModule().Source().Directory("./testdata/helm/charts/package"))
	//
	// 	registry := dag.Container().
	// 		From(fmt.Sprintf("%s:%s", zotRepository, zotVersion)).
	// 		WithExposedPort(8080).
	// 		WithMountedDirectory("/etc/zot", dag.CurrentModule().Source().Directory("./testdata/helm/zot")).
	// 		WithExec([]string{"serve", "/etc/zot/config.json"}).
	// 		AsService()
	//
	// 	password := dag.SetSecret("registry-password", "password")
	//
	// 	_, err = dag.Helm(HelmOpts{
	// 		Container: helm.Container().
	// 			WithServiceBinding("zot2", registry),
	// 	}).
	// 		Login("zot2:8080", "username", password, HelmLoginOpts{
	// 			Insecure: true,
	// 		}).
	// 		Push(ctx, pkg, "oci://zot2:8080/helm-charts", HelmPushOpts{
	// 			PlainHTTP: true,
	// 		})
	//
	// 	return err
	// })

	return group.Wait()
}

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

		helm := dag.Helm(HelmOpts{
			Version: helmVersion,
		})

		pkg := helm.Package(dag.CurrentModule().Source().Directory("./testdata/helm/charts/package"))

		registry := dag.Container().
			From(fmt.Sprintf("%s:%s", zotRepository, zotVersion)).
			WithExposedPort(8080).
			WithMountedDirectory("/etc/zot", dag.CurrentModule().Source().Directory("./testdata/helm/zot")).
			WithExec([]string{"serve", "/etc/zot/config.json"}).
			AsService()

		password := dag.SetSecret("registry-password", "password")

		_, err = dag.Helm(HelmOpts{
			Container: helm.Container().
				WithServiceBinding("zot", registry),
		}).
			Login("zot:8080", "username", password, HelmLoginOpts{
				Insecure: true,
			}).
			Push(pkg, "oci://zot:8080/helm-charts", HelmPushOpts{
				PlainHTTP: true,
			}).
			Sync(ctx)

		return err
	})

	return group.Wait()
}

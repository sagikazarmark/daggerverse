package main

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
)

type Ci struct{}

func (m *Ci) Bats() *Container {
	return dag.Bats().
		WithSource(dag.Host().Directory("./testdata/bats")).
		Run([]string{"test.bats"})
}

func (m *Ci) Go(ctx context.Context) error {
	var group errgroup.Group

	// Default container
	group.Go(func() error {
		_, err := dag.Go().
			Exec([]string{"go", "version"}).
			Sync(ctx)

		return err
	})

	// Custom version
	group.Go(func() error {
		_, err := dag.Go(GoOpts{
			Version: "latest",
		}).
			Exec([]string{"go", "version"}).
			Sync(ctx)

		return err
	})

	// Custom version (DEPRECATED)
	group.Go(func() error {
		_, err := dag.Go().
			FromVersion("latest").
			Exec([]string{"go", "version"}).
			Sync(ctx)

		return err
	})

	// Custom image
	group.Go(func() error {
		_, err := dag.Go(GoOpts{
			Image: "golang:latest",
		}).
			Exec([]string{"go", "version"}).
			Sync(ctx)

		return err
	})

	// Custom image (DEPRECATED)
	group.Go(func() error {
		_, err := dag.Go().
			FromImage("golang:latest").
			Exec([]string{"go", "version"}).
			Sync(ctx)

		return err
	})

	// Custom container
	group.Go(func() error {
		_, err := dag.Go(GoOpts{
			Container: dag.Container().From("golang:latest"),
		}).
			Exec([]string{"go", "version"}).
			Sync(ctx)

		return err
	})

	// Custom container (DEPRECATED)
	group.Go(func() error {
		_, err := dag.Go().
			FromContainer(dag.Container().From("golang:latest")).
			Exec([]string{"go", "version"}).
			Sync(ctx)

		return err
	})

	// Env vars
	group.Go(func() error {
		var group errgroup.Group

		group.Go(func() error {
			out, err := dag.Go().
				WithEnvVariable("FOO", "bar").
				Exec([]string{"bash", "-c", "echo $FOO"}).
				Stdout(ctx)
			if err != nil {
				return err
			}

			if out != "bar\n" {
				return fmt.Errorf("unexpected output: wanted \"bar\", got %q", out)
			}

			return nil
		})

		group.Go(func() error {
			out, err := dag.Go().
				FromVersion("latest").
				WithEnvVariable("FOO", "bar").
				Exec([]string{"bash", "-c", "echo $FOO"}).
				Stdout(ctx)
			if err != nil {
				return err
			}

			if out != "bar\n" {
				return fmt.Errorf("unexpected output: wanted \"bar\", got %q", out)
			}

			return nil
		})

		group.Go(func() error {
			out, err := dag.Go().
				FromVersion("latest").
				WithSource(dag.Host().Directory("./testdata/go")).
				WithEnvVariable("FOO", "bar").
				Exec([]string{"bash", "-c", "echo $FOO"}).
				Stdout(ctx)
			if err != nil {
				return err
			}

			if out != "bar\n" {
				return fmt.Errorf("unexpected output: wanted \"bar\", got %q", out)
			}

			return nil
		})

		return group.Wait()
	})

	// Platform
	group.Go(func() error {
		var group errgroup.Group

		const platform = "darwin/arm64/v7"

		group.Go(func() error {
			out, err := dag.Go().
				WithPlatform(platform).
				Exec([]string{"bash", "-c", "echo $GOOS/$GOARCH/$GOARM"}).
				Stdout(ctx)
			if err != nil {
				return err
			}

			if out != platform+"\n" {
				return fmt.Errorf("unexpected output: wanted %q, got %q", platform, out)
			}

			return nil
		})

		group.Go(func() error {
			out, err := dag.Go().
				WithSource(dag.Host().Directory("./testdata/go")).
				WithPlatform(platform).
				Exec([]string{"bash", "-c", "echo $GOOS/$GOARCH/$GOARM"}).
				Stdout(ctx)
			if err != nil {
				return err
			}

			if out != platform+"\n" {
				return fmt.Errorf("unexpected output: wanted %q, got %q", platform, out)
			}

			return nil
		})

		return group.Wait()
	})

	// CGO
	group.Go(func() error {
		var group errgroup.Group

		// Enabled
		group.Go(func() error {
			var group errgroup.Group

			group.Go(func() error {
				out, err := dag.Go().
					WithCgoEnabled().
					Exec([]string{"bash", "-c", "echo $CGO_ENABLED"}).
					Stdout(ctx)
				if err != nil {
					return err
				}

				if out != "1\n" {
					return fmt.Errorf("unexpected output: wanted \"1\", got %q", out)
				}

				return nil
			})

			group.Go(func() error {
				out, err := dag.Go().
					WithSource(dag.Host().Directory("./testdata/go")).
					WithCgoEnabled().
					Exec([]string{"bash", "-c", "echo $CGO_ENABLED"}).
					Stdout(ctx)
				if err != nil {
					return err
				}

				if out != "1\n" {
					return fmt.Errorf("unexpected output: wanted \"1\", got %q", out)
				}

				return nil
			})

			return group.Wait()
		})

		// Disabled
		group.Go(func() error {
			var group errgroup.Group

			group.Go(func() error {
				out, err := dag.Go().
					WithCgoDisabled().
					Exec([]string{"bash", "-c", "echo $CGO_ENABLED"}).
					Stdout(ctx)
				if err != nil {
					return err
				}

				if out != "0\n" {
					return fmt.Errorf("unexpected output: wanted \"0\", got %q", out)
				}

				return nil
			})

			group.Go(func() error {
				out, err := dag.Go().
					WithSource(dag.Host().Directory("./testdata/go")).
					WithCgoDisabled().
					Exec([]string{"bash", "-c", "echo $CGO_ENABLED"}).
					Stdout(ctx)
				if err != nil {
					return err
				}

				if out != "0\n" {
					return fmt.Errorf("unexpected output: wanted \"0\", got %q", out)
				}

				return nil
			})

			return group.Wait()
		})

		return group.Wait()
	})

	// Build
	group.Go(func() error {
		var group errgroup.Group

		group.Go(func() error {
			binary, err := dag.Go().
				Build(dag.Host().Directory("./testdata/go")).
				Sync(ctx)
			if err != nil {
				return err
			}

			out, err := dag.Container().From("alpine").WithFile("/app", binary).WithExec([]string{"/app"}).Stderr(ctx)
			if err != nil {
				return err
			}

			if out != "hello\n" {
				return fmt.Errorf("unexpected output: wanted \"hello\", got %q", out)
			}

			return nil
		})

		group.Go(func() error {
			binary, err := dag.Go().
				WithSource(dag.Host().Directory("./testdata/go")).
				Build().
				Sync(ctx)
			if err != nil {
				return err
			}

			out, err := dag.Container().From("alpine").WithFile("/app", binary).WithExec([]string{"/app"}).Stderr(ctx)
			if err != nil {
				return err
			}

			if out != "hello\n" {
				return fmt.Errorf("unexpected output: wanted \"hello\", got %q", out)
			}

			return nil
		})

		return group.Wait()
	})

	// Exec: Build
	group.Go(func() error {
		ctr, err := dag.Go().
			WithSource(dag.Host().Directory("./testdata/go")).
			Exec([]string{"go", "build", "-o", "/app", "."}).
			Sync(ctx)
		if err != nil {
			return err
		}

		out, err := ctr.WithExec([]string{"/app"}).Stderr(ctx)
		if err != nil {
			return err
		}

		if out != "hello\n" {
			return fmt.Errorf("unexpected output: wanted \"hello\", got %q", out)
		}

		return nil
	})

	// Exec: Test
	group.Go(func() error {
		ctr, err := dag.Go().
			WithSource(dag.Host().Directory("./testdata/go")).
			Exec([]string{"go", "test", "-v"}).
			Sync(ctx)
		if err != nil {
			return err
		}

		out, err := ctr.Stdout(ctx)
		if err != nil {
			return err
		}

		if !strings.Contains(out, "hello") {
			return fmt.Errorf("unexpected output to contain \"hello\", got %q", out)
		}

		return nil
	})

	return group.Wait()
}

func (m *Ci) GolangciLint() *Container {
	return dag.GolangciLint().
		Run(dag.Host().Directory("./testdata/go"))
}

func (m *Ci) Helm(ctx context.Context) error {
	var group errgroup.Group

	const helmVersion = "3.13.2"

	// Lint
	// TODO: improve this test
	group.Go(func() error {
		_, err := dag.Helm(HelmOpts{
			Version: helmVersion,
		}).
			Lint(dag.Host().Directory("./testdata/helm/charts/package")).
			Sync(ctx)

		return err
	})

	// Package
	// TODO: improve this test
	group.Go(func() error {
		_, err := dag.Helm(HelmOpts{
			Version: helmVersion,
		}).
			Package(dag.Host().Directory("./testdata/helm/charts/package")).
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

		pkg := helm.Package(dag.Host().Directory("./testdata/helm/charts/package"))

		registry := dag.Container().
			From(fmt.Sprintf("%s:%s", zotRepository, zotVersion)).
			WithExposedPort(8080).
			WithMountedDirectory("/etc/zot", dag.Host().Directory("./testdata/helm/zot")).
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

func (m *Ci) HelmDocs(ctx context.Context) error {
	var group errgroup.Group

	const helmDocsVersion = "v1.11.3"

	chartDir := func(chartName string) *Directory {
		return dag.Host().Directory(fmt.Sprintf("./testdata/helm-docs/charts/%s", chartName))
	}

	expected := func(chartName string) *File {
		return dag.Host().File(fmt.Sprintf("./testdata/helm-docs/charts/%s/expected.md", chartName))
	}

	testCases := []struct {
		name string
		opts HelmDocsGenerateOpts
	}{
		{
			name: "default",
		},
		{
			name: "sort-values",
			opts: HelmDocsGenerateOpts{
				SortValuesOrder: "file",
			},
		},
		{
			name: "template",
			opts: HelmDocsGenerateOpts{
				Templates: []*File{
					dag.Host().File("./testdata/helm-docs/charts/template/template.md"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		chartName := testCase.name

		group.Go(func() error {
			actual := dag.HelmDocs(HelmDocsOpts{
				Version: helmDocsVersion,
			}).
				Generate(chartDir(chartName), testCase.opts)

			_, err := dag.Container().
				From("alpine").
				WithMountedFile("/src/expected", expected(chartName)).
				WithMountedFile("/src/actual", actual).
				WithExec([]string{"diff", "-u", "/src/expected", "/src/actual"}).
				Sync(ctx)

			return err
		})
	}

	return group.Wait()
}

func (m *Ci) Kafka() *Container {
	kafka := dag.Kafka()

	return kafka.Container().
		WithServiceBinding("kafka", kafka.Service()).
		WithExec([]string{"kafka-topics.sh", "--list", "--bootstrap-server", "kafka:9092"})
}

func (m *Ci) Spectral() *Container {
	return dag.Spectral().
		WithSource(dag.Host().Directory("./testdata/spectral")).
		Lint("openapi.json")
}

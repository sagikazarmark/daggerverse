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
		_, err := dag.Go().
			FromVersion("latest").
			Exec([]string{"go", "version"}).
			Sync(ctx)

		return err
	})

	// Custom image
	group.Go(func() error {
		_, err := dag.Go().
			FromImage("golang:latest").
			Exec([]string{"go", "version"}).
			Sync(ctx)

		return err
	})

	// Custom container
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

	// Build
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

	// Test
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
		Run(GolangciLintRunOpts{
			Source: dag.Host().Directory("./testdata/go"),
		})
}

func (m *Ci) HelmDocs(ctx context.Context) error {
	var group errgroup.Group

	const helmDocsVersion = "1.11.3"

	chartDir := func(chartName string) *Directory {
		return dag.Host().Directory(fmt.Sprintf("./testdata/helm-docs/charts/%s", chartName))
	}

	expected := func(chartName string) *File {
		return dag.Host().File(fmt.Sprintf("./testdata/helm-docs/charts/%s/expected.md", chartName))
	}

	testCases := []string{"test"}

	for _, testCase := range testCases {
		chartName := testCase
		group.Go(func() error {
			actual := dag.HelmDocs().
				FromVersion(helmDocsVersion).
				Generate(chartName, chartDir(chartName))

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

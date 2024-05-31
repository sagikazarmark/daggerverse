package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.DefaultContainer)
	p.Go(m.CustomVersion)
	p.Go(m.CustomContainer)
	p.Go(m.EnvVars)
	p.Go(m.Platform)
	p.Go(m.Cgo)
	p.Go(m.Build)
	p.Go(m.ExecBuild)
	p.Go(m.ExecTest)

	return p.Wait()
}

func (m *Tests) DefaultContainer(ctx context.Context) error {
	_, err := dag.Go().
		Exec([]string{"go", "version"}).
		Sync(ctx)

	return err
}

func (m *Tests) CustomVersion(ctx context.Context) error {
	_, err := dag.Go(GoOpts{
		Version: "latest",
	}).
		Exec([]string{"go", "version"}).
		Sync(ctx)

	return err
}

func (m *Tests) CustomContainer(ctx context.Context) error {
	_, err := dag.Go(GoOpts{
		Container: dag.Container().From("golang:latest"),
	}).
		Exec([]string{"go", "version"}).
		Sync(ctx)

	return err
}

func (m *Tests) EnvVars(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(func(ctx context.Context) error {
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

	p.Go(func(ctx context.Context) error {
		out, err := dag.Go().
			WithSource(dag.CurrentModule().Source().Directory("./testdata")).
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

	return p.Wait()
}

func (m *Tests) Platform(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	const platform = "darwin/arm64/v7"

	p.Go(func(ctx context.Context) error {
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

	p.Go(func(ctx context.Context) error {
		out, err := dag.Go().
			WithSource(dag.CurrentModule().Source().Directory("./testdata")).
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

	return p.Wait()
}

func (m *Tests) Cgo(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	// Enabled
	p.Go(func(ctx context.Context) error {
		p := pool.New().WithErrors().WithContext(ctx)

		p.Go(func(ctx context.Context) error {
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

		p.Go(func(ctx context.Context) error {
			out, err := dag.Go().
				WithSource(dag.CurrentModule().Source().Directory("./testdata")).
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

		return p.Wait()
	})

	// Disabled
	p.Go(func(ctx context.Context) error {
		p := pool.New().WithErrors().WithContext(ctx)

		p.Go(func(ctx context.Context) error {
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

		p.Go(func(ctx context.Context) error {
			out, err := dag.Go().
				WithSource(dag.CurrentModule().Source().Directory("./testdata")).
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

		return p.Wait()
	})

	return p.Wait()
}

func (m *Tests) Build(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	const platform = "darwin/arm64/v7"

	p.Go(func(ctx context.Context) error {
		binary, err := dag.Go().
			Build(dag.CurrentModule().Source().Directory("./testdata")).
			Sync(ctx)
		if err != nil {
			return err
		}

		out, err := dag.Container().From("alpine").WithFile("/app", binary).WithExec([]string{"/app"}).Stderr(ctx)
		if err != nil {
			return err
		}

		if out != "hello" {
			return fmt.Errorf("unexpected output: wanted \"hello\", got %q", out)
		}

		return nil
	})

	p.Go(func(ctx context.Context) error {
		binary, err := dag.Go().
			WithSource(dag.CurrentModule().Source().Directory("./testdata")).
			Build().
			Sync(ctx)
		if err != nil {
			return err
		}

		out, err := dag.Container().From("alpine").WithFile("/app", binary).WithExec([]string{"/app"}).Stderr(ctx)
		if err != nil {
			return err
		}

		if out != "hello" {
			return fmt.Errorf("unexpected output: wanted \"hello\", got %q", out)
		}

		return nil
	})

	p.Go(func(ctx context.Context) error {
		binary, err := dag.Go().
			WithSource(dag.CurrentModule().Source().Directory("./testdata")).
			Build(GoWithSourceBuildOpts{
				Ldflags: []string{"-X", "main.version=1.0.0"},
			}).
			Sync(ctx)
		if err != nil {
			return err
		}

		out, err := dag.Container().From("alpine").WithFile("/app", binary).WithExec([]string{"/app", "version"}).Stderr(ctx)
		if err != nil {
			return err
		}

		if out != "1.0.0" {
			return fmt.Errorf("unexpected output: wanted \"1.0.0\", got %q", out)
		}

		return nil
	})

	return p.Wait()
}

func (m *Tests) ExecBuild(ctx context.Context) error {
	ctr, err := dag.Go().
		WithSource(dag.CurrentModule().Source().Directory("./testdata")).
		Exec([]string{"go", "build", "-o", "/app", "."}).
		Sync(ctx)
	if err != nil {
		return err
	}

	out, err := ctr.WithExec([]string{"/app"}).Stderr(ctx)
	if err != nil {
		return err
	}

	if out != "hello" {
		return fmt.Errorf("unexpected output: wanted \"hello\", got %q", out)
	}

	return nil
}

func (m *Tests) ExecTest(ctx context.Context) error {
	ctr, err := dag.Go().
		WithSource(dag.CurrentModule().Source().Directory("./testdata")).
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
}

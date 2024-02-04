package main

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func (m *Ci) Archive(ctx context.Context) error {
	var group errgroup.Group

	source := dag.CurrentModule().Source().Directory("./testdata/archive")

	// Format
	group.Go(func() error {
		var group errgroup.Group

		// tar.gz
		group.Go(func() error {
			testCases := []*File{
				// Default auto, no platform
				dag.Archive().Create("test", source),

				// Default auto on Linux
				dag.Archive().Create("test", source, ArchiveCreateOpts{
					Format:   "auto",
					Platform: "linux/amd64",
				}),

				// tar.gz
				dag.Archive().Create("test", source, ArchiveCreateOpts{
					Format:   "tar.gz",
					Platform: "whatever",
				}),
			}

			var group errgroup.Group

			for _, testCase := range testCases {
				testCase := testCase

				group.Go(func() error {
					// TODO: check file name once Dagger 0.9.7 is released

					_, err := dag.Container().From("alpine:latest").
						WithWorkdir("/work").
						WithMountedFile("/work/test.tar.gz", testCase).
						WithExec([]string{"tar", "-xzf", "test.tar.gz"}).
						WithExec([]string{"sh", "-c", "test -f hello"}).
						WithExec([]string{"sh", "-c", "test -d foo"}).
						WithExec([]string{"sh", "-c", "test -f foo/bar"}).
						Sync(ctx)

					return err
				})
			}

			return group.Wait()
		})

		// zip
		group.Go(func() error {
			testCases := []*File{
				// Default auto on Windows
				dag.Archive().Create("test", source, ArchiveCreateOpts{
					Platform: "windows/amd64",
				}),

				// Auto on Windows
				dag.Archive().Create("test", source, ArchiveCreateOpts{
					Format:   "auto",
					Platform: "windows/amd64",
				}),

				// zip
				dag.Archive().Create("test", source, ArchiveCreateOpts{
					Format:   "zip",
					Platform: "whatever",
				}),
			}

			var group errgroup.Group

			for _, testCase := range testCases {
				testCase := testCase

				group.Go(func() error {
					// TODO: check file name once Dagger 0.9.7 is released

					_, err := dag.Container().From("alpine:latest").
						WithWorkdir("/work").
						WithMountedFile("/work/test.zip", testCase).
						WithExec([]string{"unzip", "test.zip"}).
						WithExec([]string{"sh", "-c", "test -f hello"}).
						WithExec([]string{"sh", "-c", "test -d foo"}).
						WithExec([]string{"sh", "-c", "test -f foo/bar"}).
						Sync(ctx)

					return err
				})
			}

			return group.Wait()
		})

		return group.Wait()
	})

	return group.Wait()
}

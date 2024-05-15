package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.ArchiveFiles)
	p.Go(m.ArchiveDirectory)

	return p.Wait()
}

func (m *Tests) ArchiveFiles(ctx context.Context) error {
	dir := dag.CurrentModule().Source().Directory("./testdata")

	archive := dag.Arc().
		ArchiveFiles(
			"test",
			[]*File{
				dir.File("hello"),
				dir.File("foo/bar"),
			},
		).
		TarGz()

	unarchivedDir := dag.Arc().Unarchive(archive)

	entries, err := unarchivedDir.Directory("test").Entries(ctx)
	if err != nil {
		return err
	}

	if !slices.Equal(entries, []string{"bar", "hello"}) {
		return fmt.Errorf("unexpected entries: %v", entries)
	}

	return nil
}

func (m *Tests) ArchiveDirectory(ctx context.Context) error {
	dir := dag.CurrentModule().Source().Directory("./testdata")

	archive := dag.Arc().ArchiveDirectory("test", dir).TarGz()

	unarchivedDir := dag.Arc().Unarchive(archive)

	entries, err := unarchivedDir.Directory("test").Entries(ctx)
	if err != nil {
		return err
	}

	if !slices.Equal(entries, []string{"foo", "hello"}) {
		return fmt.Errorf("unexpected entries: %v", entries)
	}

	return nil
}

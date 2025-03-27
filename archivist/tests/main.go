package main

import (
	"context"
	"dagger/archivist/tests/internal/dagger"
	"fmt"
	"slices"

	"github.com/sourcegraph/conc/pool"
)

type Tests struct{}

// All executes all tests.
func (m *Tests) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Tar)
	p.Go(m.TarBr)
	p.Go(m.TarBz2)
	p.Go(m.TarGz)
	// p.Go(m.TarLz4) // breaking: "creating tar: wrapping writer: lz4: invalid compression level: 128"
	p.Go(m.TarSz)
	p.Go(m.TarXz)
	p.Go(m.TarZst)
	p.Go(m.Zip)

	return p.Wait()
}

func (m *Tests) Tar(ctx context.Context) error {
	return test(ctx, dag.Archivist().Tar())
}

func (m *Tests) TarBr(ctx context.Context) error {
	return test(ctx, dag.Archivist().TarBr())
}

func (m *Tests) TarBz2(ctx context.Context) error {
	return test(ctx, dag.Archivist().TarBz2())
}

func (m *Tests) TarGz(ctx context.Context) error {
	return test(ctx, dag.Archivist().TarGz())
}

func (m *Tests) TarLz4(ctx context.Context) error {
	return test(ctx, dag.Archivist().TarLz4())
}

func (m *Tests) TarSz(ctx context.Context) error {
	return test(ctx, dag.Archivist().TarSz())
}

func (m *Tests) TarXz(ctx context.Context) error {
	return test(ctx, dag.Archivist().TarXz())
}

func (m *Tests) TarZst(ctx context.Context) error {
	return test(ctx, dag.Archivist().TarZst())
}

func (m *Tests) Zip(ctx context.Context) error {
	return test(ctx, dag.Archivist().Zip())
}

type archiver interface {
	Archive(name string, source *dagger.Directory) *dagger.File
}

func test(ctx context.Context, a archiver) error {
	dir := dag.CurrentModule().Source().Directory("./testdata")

	archive := a.Archive("test", dir)

	unarchivedDir := dag.Arc().Unarchive(archive)

	entries, err := unarchivedDir.Directory("test").Entries(ctx)
	if err != nil {
		return err
	}

	if !slices.Equal(entries, []string{"foo/", "hello"}) {
		return fmt.Errorf("unexpected entries: %v", entries)
	}

	return nil
}

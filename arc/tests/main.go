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

	p.Go(m.ArchiveFiles().All)
	p.Go(m.ArchiveDirectory().All)

	return p.Wait()
}

func (m *Tests) ArchiveFiles() *ArchiveFiles {
	return &ArchiveFiles{}
}

type ArchiveFiles struct{}

// All executes all tests.
func (m *ArchiveFiles) All(ctx context.Context) error {
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

func (m *ArchiveFiles) Tar(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.Tar() })
}

func (m *ArchiveFiles) TarBr(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarBr() })
}

func (m *ArchiveFiles) TarBz2(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarBz2() })
}

func (m *ArchiveFiles) TarGz(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarGz() })
}

func (m *ArchiveFiles) TarLz4(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarLz4() })
}

func (m *ArchiveFiles) TarSz(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarSz() })
}

func (m *ArchiveFiles) TarXz(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarXz() })
}

func (m *ArchiveFiles) TarZst(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarZst() })
}

func (m *ArchiveFiles) Zip(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.Zip() })
}

func (m *ArchiveFiles) test(ctx context.Context, callback func(*ArcArchive) *File) error {
	dir := dag.CurrentModule().Source().Directory("./testdata")

	archive := callback(
		dag.Arc().
			ArchiveFiles(
				"test",
				[]*File{
					dir.File("hello"),
					dir.File("foo/bar"),
				},
			),
	)

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

func (m *Tests) ArchiveDirectory() *ArchiveDirectory {
	return &ArchiveDirectory{}
}

type ArchiveDirectory struct{}

// All executes all tests.
func (m *ArchiveDirectory) All(ctx context.Context) error {
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

func (m *ArchiveDirectory) Tar(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.Tar() })
}

func (m *ArchiveDirectory) TarBr(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarBr() })
}

func (m *ArchiveDirectory) TarBz2(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarBz2() })
}

func (m *ArchiveDirectory) TarGz(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarGz() })
}

func (m *ArchiveDirectory) TarLz4(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarLz4() })
}

func (m *ArchiveDirectory) TarSz(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarSz() })
}

func (m *ArchiveDirectory) TarXz(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarXz() })
}

func (m *ArchiveDirectory) TarZst(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.TarZst() })
}

func (m *ArchiveDirectory) Zip(ctx context.Context) error {
	return m.test(ctx, func(a *ArcArchive) *File { return a.Zip() })
}

func (m *ArchiveDirectory) test(ctx context.Context, callback func(*ArcArchive) *File) error {
	dir := dag.CurrentModule().Source().Directory("./testdata")

	archive := callback(dag.Arc().ArchiveDirectory("test", dir))

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

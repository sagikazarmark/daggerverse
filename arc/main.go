// Easily create & extract archives, and compress & decompress files of various formats.
package main

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

const (
	alpineBaseImage = "alpine:latest"
	latestVersion   = "3.5.0" // latest version with published binaries
)

// Easily create & extract archives, and compress & decompress files of various formats.
type Arc struct {
	// +private
	Ctr *Container
}

func New(
	// Version to download from GitHub Releases (default: "3.5.0").
	// +optional
	version string,

	// Custom image reference in "repository:tag" format to use as a base container.
	// +optional
	image string,

	// Custom container to use as a base container.
	// +optional
	container *Container,
) *Arc {
	var ctr *Container

	if version != "" {
		ctr = containerWithDownloadedBinary(version)
	} else if image != "" {
		ctr = dag.Container().From(image)
	} else if container != nil {
		ctr = container
	} else {
		ctr = containerWithDownloadedBinary(latestVersion)
	}

	return &Arc{ctr}
}

func containerWithDownloadedBinary(version string) *Container {
	return dag.Container().
		From(alpineBaseImage).
		WithExec([]string{"wget", "-O", "/usr/local/bin/arc", fmt.Sprintf("https://github.com/mholt/archiver/releases/download/v%s/arc_%s_%s_%s", version, version, runtime.GOOS, runtime.GOARCH)}).
		WithExec([]string{"chmod", "+x", "/usr/local/bin/arc"})
}

// Create a new archive from a list of files.
func (m *Arc) ArchiveFiles(
	// File name without extension.
	name string,

	// Files to archive.
	files []*File,
) *Archive {
	dir := dag.Directory()
	for _, file := range files {
		dir = dir.WithFile("", file)
	}

	return &Archive{
		Name:      name,
		Directory: dir,
		Ctr:       m.Ctr,
	}
}

// Create a new archive from the contents of a directory.
func (m *Arc) ArchiveDirectory(
	// File name without extension.
	name string,

	// Directory to archive.
	directory *Directory,
) *Archive {
	return &Archive{
		Name:      name,
		Directory: directory,
		Ctr:       m.Ctr,
	}
}

type Archive struct {
	// File name of the archive (without extension).
	Name string

	// +private
	Directory *Directory

	// +private
	Ctr *Container
}

var supportedFormats = []string{
	"tar",
	"tar.br", "tbr",
	"tar.bz2", "tbz2",
	"tar.gz", "tgz",
	"tar.lz4", "tlz4",
	"tar.sz", "tsz",
	"tar.xz", "txz",
	"tar.zst",
	"zip",
}

// Create an archive from the provided files or directory.
func (m *Archive) Create(
	// One of the supported archive formats. (choices: "zip", "tar", "tar.br", "tbr", "tar.gz", "tgz", "tar.bz2", "tbz2", "tar.xz", "txz", "tar.lz4", "tlz4", "tar.sz", "tsz", "tar.zst")
	format string,
) (*File, error) {
	format = strings.ToLower(format)

	if !slices.Contains(supportedFormats, format) {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	archiveFilePath := path.Join("/work", m.Name+"."+format)
	cmd := []string{"arc", "-folder-safe=false", "archive", archiveFilePath, "$(ls)"}

	return m.Ctr.
		WithWorkdir("/work/src").
		WithMountedDirectory("/work/src", m.Directory).
		WithExec([]string{"sh", "-c", strings.Join(cmd, " ")}).
		File(archiveFilePath), nil
}

func (m *Archive) Tar() (*File, error)    { return m.Create("tar") }
func (m *Archive) TarBr() (*File, error)  { return m.Create("tar.br") }
func (m *Archive) TarBz2() (*File, error) { return m.Create("tar.bz2") }
func (m *Archive) TarGz() (*File, error)  { return m.Create("tar.gz") }
func (m *Archive) TarLz4() (*File, error) { return m.Create("tar.lz4") }
func (m *Archive) TarSz() (*File, error)  { return m.Create("tar.sz") }
func (m *Archive) TarXz() (*File, error)  { return m.Create("tar.xz") }
func (m *Archive) TarZst() (*File, error) { return m.Create("tar.zst") }
func (m *Archive) Zip() (*File, error)    { return m.Create("zip") }

// Extract the contents of an archive.
func (m *Arc) Unarchive(
	ctx context.Context,

	// Archive file (in one of the supported formats).
	archive *File,
) (*Directory, error) {
	fileName, err := archive.Name(ctx)
	if err != nil {
		return nil, err
	}

	baseName := trimExt(fileName)
	destination := path.Join("/work", baseName)

	cmd := []string{"arc", "unarchive", fileName, baseName}

	return m.Ctr.
		WithWorkdir("/work").
		WithMountedFile(path.Join("/work", fileName), archive).
		WithExec([]string{"sh", "-c", strings.Join(cmd, " ")}).
		Directory(destination), nil
}

func trimExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

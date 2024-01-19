package main

import (
	"fmt"
	"path"
	"strings"

	"golang.org/x/exp/slices"
)

var supportedFormats = []string{"tar.gz", "zip"}

type Archive struct{}

// Create an archive from a directory of files.
func (m *Archive) Create(
	// File name without extension.
	name string,

	// The directory to archive.
	source *Directory,

	// Archive format. (choices: "auto", "tar.gz", "zip") (default "auto")
	//
	// "auto" will attempt to choose the best format. If platform is specified and the platform is Windows, "zip" will be used.
	// Otherwise, "tar.gz" will be used.
	// +optional
	format string,

	// Target platform in "[os]/[platform]/[version]" format (e.g., "darwin/arm64/v7", "windows/amd64", "linux/arm64").
	// +optional
	platform Platform,
) (*File, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	format, err := processFormat(format, platform)
	if err != nil {
		return nil, err
	}

	fileName := name + "." + format

	return dag.Container().
		From("alpine:latest").
		WithWorkdir("/work").
		WithMountedDirectory("/work/src", source).
		With(func(c *Container) *Container {
			switch format {
			case "tar.gz":
				c = c.WithExec([]string{"tar", "-czf", fileName, "-C", "./src", "."})
			case "zip":
				c = c.
					WithExec([]string{"apk", "add", "--update", "--no-cache", "zip"}).
					WithWorkdir("/work/src").
					WithExec([]string{"zip", path.Join("..", fileName), "-r", "."})
			}

			return c
		}).
		File(path.Join("/work", fileName)), nil
}

func processFormat(format string, platform Platform) (string, error) {
	format = strings.ToLower(format)

	if format == "" {
		format = "auto"
	}

	if format == "auto" {
		if platform != "" && strings.HasPrefix(string(platform), "windows/") {
			format = "zip"
		} else {
			format = "tar.gz"
		}
	}

	if !slices.Contains(supportedFormats, format) {
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	return format, nil
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "stoplight/spectral"

type Spectral struct {
	// +private
	Ctr *Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	// +optional
	version string,

	// Custom image reference in "repository:tag" format to use as a base container.
	// +optional
	image string,

	// Custom container to use as a base container.
	// +optional
	container *Container,
) *Spectral {
	var ctr *Container

	if version != "" {
		ctr = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	} else if image != "" {
		ctr = dag.Container().From(image)
	} else if container != nil {
		ctr = container
	} else {
		ctr = dag.Container().From(defaultImageRepository)
	}

	return &Spectral{ctr}
}

func (m *Spectral) Container() *Container {
	return m.Ctr
}

// Lint JSON/YAML documents.
func (m *Spectral) Lint(
	ctx context.Context,

	// JSON/YAML OpenAPI documents.
	documents []*File,

	// Ruleset file.
	ruleset *File,

	// Results of this level or above will trigger a failure exit code. (choices: "error", "warn", "info", "hint") (default "error")
	// +optional
	failSeverity string,

	// Only output results equal to or greater than fail severity.
	// +optional
	displayOnlyFailures bool,

	// Custom json-ref-resolver instance.
	// +optional
	resolver *File,

	// Text encoding to use. (choices: "utf8", "ascii", "utf-8", "utf16le", "ucs2", "ucs-2", "base64", "latin1") (default "utf8")
	// +optional
	encoding string,

	// Increase verbosity.
	// +optional
	verbose bool,

	// No logging, output only.
	// +optional
	quiet bool,
) (*Container, error) {
	ctr := m.Ctr
	args := []string{"lint"}

	rulesetType, err := detectRulesetType(ctx, ruleset)
	if err != nil {
		return nil, err
	}

	rulesetFilePath := fmt.Sprintf("/work/ruleset.%s", rulesetType)
	ctr = ctr.WithMountedFile(rulesetFilePath, ruleset)
	args = append(args, "--ruleset", rulesetFilePath)

	if failSeverity != "" {
		args = append(args, "--fail-severity", failSeverity)
	}

	if resolver != nil {
		ctr = ctr.WithMountedFile("/work/resolver", resolver)
		args = append(args, "--resolver", "/work/resolver")
	}

	if verbose {
		args = append(args, "--verbose")
	}

	if quiet {
		args = append(args, "--quiet")
	}

	for i, document := range documents {
		documentName := fmt.Sprintf("/work/documents/document-%d", i)

		ctr = ctr.WithMountedFile(documentName, document)
		args = append(args, documentName)
	}

	return ctr.WithExec(args), nil
}

// This is a workaround for the fact that Dagger does not preserve file names.
// https://github.com/dagger/dagger/issues/6416
func detectRulesetType(ctx context.Context, ruleset *File) (string, error) {
	rulesetContents, err := ruleset.Contents(ctx)
	if err != nil {
		return "", err
	}

	// Fall back to JS type
	rulesetType := "js"

	if err := yaml.Unmarshal([]byte(rulesetContents), io.Discard); err == nil {
		rulesetType = "yaml"
	} else if err = json.Unmarshal([]byte(rulesetContents), io.Discard); err == nil {
		rulesetType = "json"
	}

	return rulesetType, nil
}

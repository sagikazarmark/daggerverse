package main

import (
	"context"
	"dagger/helm/internal/dagger"
	"errors"
	"path/filepath"
	"strings"
	"time"
)

// Install a Helm chart.
func (c *Chart) Install(
	ctx context.Context,

	// Helm release name.
	name string,

	// If set, the installation process deletes the installation on failure. Wait flag will be set automatically if atomic is used.
	//
	// +optional
	atomic bool,

	// Verify certificates of HTTPS-enabled servers using this CA bundle.
	//
	// +optional
	caFile *dagger.File,

	// Identify HTTPS client using this SSL certificate file.
	//
	// +optional
	certFile *dagger.File,

	// Create the release namespace if not present.
	//
	// +optional
	createNamespace bool,

	// Update dependencies if they are missing before installing the chart.
	//
	// +optional
	dependencyUpdate bool,

	// Add a custom description.
	//
	// +optional
	description string,

	// If set, the installation process will not validate rendered templates against the Kubernetes OpenAPI Schema.
	//
	// +optional
	disableOpenapiValidation bool,

	// simulate an install. If --dry-run is set with no option being specified or as '--dry-run=client', it will not attempt cluster connections. Setting '--dry-run=server' allows attempting cluster connections..
	//
	// +optional
	// dryRun bool,

	// Enable DNS lookups when rendering templates.
	//
	// +optional
	enableDns bool,

	// Force resource updates through a replacement strategy.
	//
	// +optional
	force bool,

	// Generate the name.
	//
	// +optional
	generateName bool,

	// Hide Kubernetes Secrets when also using dry run.
	//
	// +optional
	// hideSecret bool,

	// Skip tls certificate checks for the chart download.
	//
	// +optional
	insecureSkipTlsVerify bool,

	// Identify HTTPS client using this SSL key file.
	//
	// +optional
	keyFile *dagger.Secret,

	// Labels that would be added to release metadata.
	//
	// +optional
	labels []string,

	// Specify template used to name the release.
	//
	// +optional
	nameTemplate string,

	// Prevent hooks from running during install.
	//
	// +optional
	noHooks bool,

	// output
	// passCredentials
	// password

	// Use insecure HTTP connections for the chart download.
	//
	// +optional
	plainHttp bool,

	// The path to an executable to be used for post rendering. If it exists in $PATH, the binary will be used, otherwise it will try to look for the executable at the given path.
	//
	// +optional
	postRenderer string,

	// Arguments to the post-renderer.
	//
	// +optional
	postRendererArgs []string,

	// If set, render subchart notes along with the parent.
	//
	// +optional
	renderSubchartNotes bool,

	// Re-use the given name, only if that name is a deleted release which remains in the history. This is unsafe in production.
	//
	// +optional
	replace bool,

	// repo

	// set
	// setFile
	// setJson
	// setLiteral
	// setString

	// If set, no CRDs will be installed. By default, CRDs are installed if not already present.
	//
	// +optional
	skipCrds bool,

	// Time to wait for any individual Kubernetes operation (like Jobs for hooks).
	//
	// +optional
	timeout string,

	// username

	// Specify values in a YAML file.
	//
	// +optional
	values []*dagger.File,

	// Verify the package before using it.
	//
	// +optional
	verify bool,

	// If set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment, StatefulSet, or ReplicaSet are in a ready state before marking the release as successful. It will wait for as long as timeout.
	//
	// +optional
	wait bool,

	// If set and wait enabled, will wait until all Jobs have been completed before marking the release as successful. It will wait for as long as timeout.
	//
	// +optional
	waitForJobs bool,
) (*Release, error) {
	chartMetadata, err := getChartMetadata(ctx, c.Directory)
	if err != nil {
		return nil, err
	}

	chartName := chartMetadata.Name
	chartPath := filepath.Join("/work/chart", chartName)

	container := c.Helm.container().
		WithMountedDirectory(chartPath, c.Directory).
		WithDirectory("/work/values", dag.Directory())

	return install(
		ctx,

		c.Helm,

		name,
		chartPath,
		container,

		atomic,
		caFile,
		certFile,
		createNamespace,
		dependencyUpdate,
		description,
		disableOpenapiValidation,
		enableDns,
		force,
		generateName,
		insecureSkipTlsVerify,
		keyFile,
		labels,
		nameTemplate,
		noHooks,
		plainHttp,
		postRenderer,
		postRendererArgs,
		renderSubchartNotes,
		replace,
		skipCrds,
		timeout,
		values,
		verify,
		wait,
		waitForJobs,
	)
}

// Install a Helm chart.
func (p *Package) Install(
	ctx context.Context,

	// Helm release name.
	name string,

	// If set, the installation process deletes the installation on failure. Wait flag will be set automatically if atomic is used.
	//
	// +optional
	atomic bool,

	// Verify certificates of HTTPS-enabled servers using this CA bundle.
	//
	// +optional
	caFile *dagger.File,

	// Identify HTTPS client using this SSL certificate file.
	//
	// +optional
	certFile *dagger.File,

	// Create the release namespace if not present.
	//
	// +optional
	createNamespace bool,

	// Update dependencies if they are missing before installing the chart.
	//
	// +optional
	dependencyUpdate bool,

	// Add a custom description.
	//
	// +optional
	description string,

	// If set, the installation process will not validate rendered templates against the Kubernetes OpenAPI Schema.
	//
	// +optional
	disableOpenapiValidation bool,

	// simulate an install. If --dry-run is set with no option being specified or as '--dry-run=client', it will not attempt cluster connections. Setting '--dry-run=server' allows attempting cluster connections..
	//
	// +optional
	// dryRun bool,

	// Enable DNS lookups when rendering templates.
	//
	// +optional
	enableDns bool,

	// Force resource updates through a replacement strategy.
	//
	// +optional
	force bool,

	// Generate the name.
	//
	// +optional
	generateName bool,

	// Hide Kubernetes Secrets when also using dry run.
	//
	// +optional
	// hideSecret bool,

	// Skip tls certificate checks for the chart download.
	//
	// +optional
	insecureSkipTlsVerify bool,

	// Identify HTTPS client using this SSL key file.
	//
	// +optional
	keyFile *dagger.Secret,

	// Labels that would be added to release metadata.
	//
	// +optional
	labels []string,

	// Specify template used to name the release.
	//
	// +optional
	nameTemplate string,

	// Prevent hooks from running during install.
	//
	// +optional
	noHooks bool,

	// output
	// passCredentials
	// password

	// Use insecure HTTP connections for the chart download.
	//
	// +optional
	plainHttp bool,

	// The path to an executable to be used for post rendering. If it exists in $PATH, the binary will be used, otherwise it will try to look for the executable at the given path.
	//
	// +optional
	postRenderer string,

	// Arguments to the post-renderer.
	//
	// +optional
	postRendererArgs []string,

	// If set, render subchart notes along with the parent.
	//
	// +optional
	renderSubchartNotes bool,

	// Re-use the given name, only if that name is a deleted release which remains in the history. This is unsafe in production.
	//
	// +optional
	replace bool,

	// repo

	// set
	// setFile
	// setJson
	// setLiteral
	// setString

	// If set, no CRDs will be installed. By default, CRDs are installed if not already present.
	//
	// +optional
	skipCrds bool,

	// Time to wait for any individual Kubernetes operation (like Jobs for hooks).
	//
	// +optional
	timeout string,

	// username

	// Specify values in a YAML file.
	//
	// +optional
	values []*dagger.File,

	// Verify the package before using it.
	//
	// +optional
	verify bool,

	// If set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment, StatefulSet, or ReplicaSet are in a ready state before marking the release as successful. It will wait for as long as timeout.
	//
	// +optional
	wait bool,

	// If set and wait enabled, will wait until all Jobs have been completed before marking the release as successful. It will wait for as long as timeout.
	//
	// +optional
	waitForJobs bool,
) (*Release, error) {
	if p.Chart == nil {
		return nil, errors.New("chart is unavailable")
	}

	chartFileName, err := p.File.Name(ctx)
	if err != nil {
		return nil, err
	}

	chartPath := filepath.Join("/work/chart", chartFileName)

	container := p.Helm.container().
		WithMountedFile(chartPath, p.File).
		WithDirectory("/work/values", dag.Directory())

	return install(
		ctx,

		p.Helm,

		name,
		chartPath,
		container,

		atomic,
		caFile,
		certFile,
		createNamespace,
		dependencyUpdate,
		description,
		disableOpenapiValidation,
		enableDns,
		force,
		generateName,
		insecureSkipTlsVerify,
		keyFile,
		labels,
		nameTemplate,
		noHooks,
		plainHttp,
		postRenderer,
		postRendererArgs,
		renderSubchartNotes,
		replace,
		skipCrds,
		timeout,
		values,
		verify,
		wait,
		waitForJobs,
	)
}

// Install a Helm chart.
func install(
	ctx context.Context,

	helm *Helm,

	name string,
	chartPath string,
	container *dagger.Container,

	atomic bool,
	caFile *dagger.File,
	certFile *dagger.File,
	createNamespace bool,
	dependencyUpdate bool,
	description string,
	disableOpenapiValidation bool,
	// dryRun bool,
	enableDns bool,
	force bool,
	generateName bool,
	// hideSecret bool,
	insecureSkipTlsVerify bool,
	keyFile *dagger.Secret,
	labels []string,
	nameTemplate string,
	noHooks bool,
	// output
	// passCredentials
	// password
	plainHttp bool,
	postRenderer string,
	postRendererArgs []string,
	renderSubchartNotes bool,
	replace bool,
	// repo
	// set
	// setFile
	// setJson
	// setLiteral
	// setString
	skipCrds bool,
	timeout string,
	// username
	values []*dagger.File,
	verify bool,
	wait bool,
	waitForJobs bool,
) (*Release, error) {
	args := []string{"helm", "install", name, chartPath}

	if atomic {
		args = append(args, "--atomic")
	}

	if caFile != nil {
		container = container.WithMountedFile("/etc/helm/ca.pem", caFile)
		args = append(args, "--ca-file", "/etc/helm/ca.pem")
	}

	if certFile != nil {
		container = container.WithMountedFile("/etc/helm/cert.pem", certFile)
		args = append(args, "--cert-file", "/etc/helm/cert.pem")
	}

	if createNamespace {
		args = append(args, "--create-namespace")
	}

	if dependencyUpdate {
		args = append(args, "--dependency-update")
	}

	if description != "" {
		args = append(args, "--description", description)
	}

	if disableOpenapiValidation {
		args = append(args, "--disable-openapi-validation")
	}

	if enableDns {
		args = append(args, "--enable-dns")
	}

	if force {
		args = append(args, "--force")
	}

	if generateName {
		args = append(args, "--generate-name")
	}

	if insecureSkipTlsVerify {
		args = append(args, "--insecure-skip-tls-verify")
	}

	if keyFile != nil {
		container = container.WithMountedSecret("/etc/helm/key.pem", keyFile)
		args = append(args, "--key-file", "/etc/helm/key.pem")
	}

	if len(labels) > 0 {
		args = append(args, "--label", strings.Join(labels, ","))
	}

	if nameTemplate != "" {
		args = append(args, "--name-template", nameTemplate)
	}

	if noHooks {
		args = append(args, "--no-hooks")
	}

	if plainHttp {
		args = append(args, "--plain-http")
	}

	if postRenderer != "" {
		args = append(args, "--post-renderer", postRenderer)
	}

	for _, postRendererArg := range postRendererArgs {
		args = append(args, "--post-renderer-args", postRendererArg)
	}

	if renderSubchartNotes {
		args = append(args, "--render-subchart-notes")
	}

	if replace {
		args = append(args, "--replace")
	}

	if skipCrds {
		args = append(args, "--skip-crds")
	}

	if timeout != "" {
		args = append(args, "--timeout", timeout)
	}

	for _, file := range values {
		name, err := file.Name(ctx)
		if err != nil {
			return nil, err
		}

		path := filepath.Join("/work/values", name)

		container = container.WithMountedFile(path, file)
		args = append(args, "--values", path)
	}

	if verify {
		args = append(args, "--verify")
	}

	if wait {
		args = append(args, "--wait")

		if waitForJobs {
			args = append(args, "--wait-for-jobs")
		}
	}

	container, err := container.WithExec(args).Sync(ctx)
	if err != nil {
		return nil, err
	}

	return &Release{
		Name:      name,
		Container: helm.container(),
	}, nil
}

type Release struct {
	Name string

	// private
	Container *dagger.Container
}

// Run Helm tests.
func (r *Release) Test(
	ctx context.Context,

	// Specify tests by attribute (currently "name") using attribute=value syntax or '!attribute=value' to exclude a test.
	//
	// +optional
	filter []string,

	// Dump the logs from test pods (this runs after all tests are complete, but before any cleanup).
	//
	// +optional
	logs bool,

	// Time to wait for any individual Kubernetes operation (like Jobs for hooks) (default 5m0s).
	//
	// +optional
	timeout string,
) (string, error) {
	args := []string{"helm", "test", r.Name}

	if len(filter) > 0 {
		args = append(args, "--filter", strings.Join(filter, ","))
	}

	if logs {
		args = append(args, "--logs")
	}

	if timeout != "" {
		_, err := time.ParseDuration(timeout)
		if err != nil {
			return "", err
		}

		args = append(args, "--timeout", timeout)
	}

	return r.Container.WithExec(args).Stdout(ctx)
}

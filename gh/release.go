package main

import (
	"context"
	"path"
)

// Manage releases.
func (m *Gh) Release() *Release {
	return &Release{Gh: m}
}

type Release struct {
	// +private
	Gh *Gh
}

// Create a new GitHub Release for a repository.
func (m *Release) Create(
	ctx context.Context,

	// Tag this release should point to or create.
	tag string,

	// Release title.
	title string,

	// Release assets to upload.
	// +optional
	files []*File,

	// Save the release as a draft instead of publishing it.
	// +optional
	draft bool,

	// Mark the release as a prerelease.
	// +optional
	preRelease bool,

	// Target branch or full commit SHA (default: main branch).
	// +optional
	target string,

	// Release notes.
	// +optional
	notes string,

	// Read release notes from file.
	// +optional
	notesFile *File,

	// Start a discussion in the specified category.
	// +optional
	discussionCategory string,

	// Automatically generate title and notes for the release.
	// +optional
	generateNotes bool,

	// Tag to use as the starting point for generating release notes.
	// +optional
	notesStartTag string,

	// Mark this release as "Latest" (default: automatic based on date and version).
	// +optional
	latest bool,

	// Abort in case the git tag doesn't already exist in the remote repository.
	// +optional
	verifyTag bool,

	// Tag to use as the starting point for generating release notes.
	// +optional
	notesFromTag bool,

	// GitHub token.
	// +optional
	token *Secret,

	// GitHub repository (e.g. "owner/repo").
	// +optional
	repo string,
) (*Container, error) {
	ctr, err := m.Gh.container(ctx, token, repo)
	if err != nil {
		return nil, err
	}

	args := []string{
		"release", "create",

		"--title", title,
	}

	if draft {
		args = append(args, "--draft")
	}

	if preRelease {
		args = append(args, "--prerelease")
	}

	if target != "" {
		args = append(args, "--target", target)
	}

	if notes != "" {
		args = append(args, "--notes", notes)
	}

	if notesFile != nil {
		ctr.WithMountedFile("/work/notes.md", notesFile)
		args = append(args, "--notes-file", "/work/notes.md")
	}

	if discussionCategory != "" {
		args = append(args, "--discussion-category", discussionCategory)
	}

	if generateNotes {
		args = append(args, "--generate-notes")
	}

	if notesStartTag != "" {
		args = append(args, "--notes-start-tag", notesStartTag)
	}

	if latest {
		args = append(args, "--latest")
	}

	if verifyTag {
		args = append(args, "--verify-tag")
	}

	if notesFromTag {
		args = append(args, "--notes-from-tag")
	}

	args = append(args, tag)

	{
		dir := dag.Directory()

		for _, file := range files {
			dir = dir.WithFile("", file)
		}

		entries, err := dir.Entries(ctx)
		if err != nil {
			return nil, err
		}

		ctr = ctr.WithMountedDirectory("/work/assets", dir)

		for _, e := range entries {
			args = append(args, path.Join("/work/assets", e))
		}
	}

	return ctr.WithExec(args), nil
}

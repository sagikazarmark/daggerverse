package main

import "context"

// Work with GitHub pull requests.
func (m *Gh) PullRequest() *PullRequest {
	return &PullRequest{Gh: m}
}

type PullRequest struct {
	// +private
	Gh *Gh
}

// Add a review to a pull request.
func (m *PullRequest) Review(
	// Pull request number, url or branch name.
	pullRequest string,

	// Specify the body of a review.
	//
	// +optional
	body string,

	// Read body text from file.
	//
	// +optional
	bodyFile *File,
) *PullRequestReview {
	return &PullRequestReview{
		PullRequest: pullRequest,
		Body:        body,
		BodyFile:    bodyFile,
		Gh:          m.Gh,
	}
}

// TODO: revisit if these should be private
type PullRequestReview struct {
	// +private
	PullRequest string

	// +private
	Body string

	// +private
	BodyFile *File

	// +private
	Gh *Gh
}

// Approve a pull request.
func (m *PullRequestReview) Approve(ctx context.Context) error {
	return m.do(ctx, "approve")
}

// Comment on a pull request.
func (m *PullRequestReview) Comment(ctx context.Context) error {
	return m.do(ctx, "comment")
}

// Request changes on a pull request.
func (m *PullRequestReview) RequestChanges(ctx context.Context) error {
	return m.do(ctx, "request-changes")
}

// Request changes on a pull request.
func (m *PullRequestReview) do(ctx context.Context, action string) error {
	args := []string{"gh", "pr", "review", m.PullRequest, "--" + action}

	_, err := m.Gh.container(nil, "").
		With(func(c *Container) *Container {
			if m.Body != "" {
				args = append(args, "--body", m.Body)
			}

			if m.BodyFile != nil {
				const bodyFilePath = "/work/tmp/body"

				c = c.WithMountedFile(bodyFilePath, m.BodyFile)

				args = append(args, "--body-file", bodyFilePath)
			}

			return c
		}).
		WithExec(args).
		Sync(ctx)

	return err
}

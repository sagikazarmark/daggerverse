// Stainless is an API SDK generator tool.
package main

import (
	"bytes"
	"context"
	"dagger/stainless/internal/dagger"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type Stainless struct {
	// +private
	Token *dagger.Secret
}

func New(token *dagger.Secret) *Stainless {
	return &Stainless{
		Token: token,
	}
}

func (m *Stainless) UploadSpec(
	ctx context.Context,

	// Stainless project name.
	projectName string,

	// OpenAPI spec file.
	openapi *dagger.File,

	// Stainless config file.
	//
	// +optional
	config *dagger.File,

	// Commit message (following conventional commit format).
	//
	// +optional
	commitMessage string,
) (*dagger.File, error) {
	var buf bytes.Buffer

	body := multipart.NewWriter(&buf)

	err := body.WriteField("projectName", projectName)
	if err != nil {
		return nil, err
	}

	if commitMessage != "" {
		err := body.WriteField("commitMesssage", commitMessage)
		if err != nil {
			return nil, err
		}
	}

	err = writeFormFile(ctx, body, "oasSpec", openapi, "openapi.json")
	if err != nil {
		return nil, err
	}

	if config != nil {
		err = writeFormFile(ctx, body, "stainlessConfig", config, "stainless.yaml")
		if err != nil {
			return nil, err
		}
	}

	err = body.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.stainlessapi.com/api/spec", &buf)
	if err != nil {
		return nil, err
	}

	token, err := m.Token.Plaintext(ctx)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", body.FormDataContentType())

	// TODO: customize HTTP client (timeout, etc)
	client := http.DefaultClient

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(response.Body)

		return nil, fmt.Errorf("failed to upload files: %s %s", response.Status, respBody)
	}

	decoratedSpec, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return dag.Directory().WithNewFile("openapi.json", string(decoratedSpec)).File("openapi.json"), nil
}

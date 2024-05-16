package main

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"
)

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// create a new multipart form file with text/plain content type
func createFormFile(w *multipart.Writer, fieldName string, fileName string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fieldName), escapeQuotes(fileName)))
	h.Set("Content-Type", "text/plain")
	return w.CreatePart(h)
}

func writeFormFile(ctx context.Context, w *multipart.Writer, fieldName string, file *File, defaultFileName string) error {
	name, err := file.Name(ctx)
	if err != nil {
		return err
	}

	if name == "" {
		name = defaultFileName
	}

	contents, err := file.Contents(ctx)
	if err != nil {
		return err
	}

	part, err := createFormFile(w, fieldName, name)
	if err != nil {
		return err
	}

	_, err = part.Write([]byte(contents))
	if err != nil {
		return err
	}

	return nil
}

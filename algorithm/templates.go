package algorithm

import (
	"embed"
	"text/template"

	"github.com/pkg/errors"
)

//go:embed templates
var searchFS embed.FS

// GetSearchTemplates parses and returns the search templates from the
// embedded filesystem.
func GetSearchTemplates(patterns ...string) (*template.Template, error) {
	templates, err := template.ParseFS(searchFS, patterns...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse search templates")
	}

	return templates, nil
}

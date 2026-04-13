package templates

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"text/template"
)

//go:embed *.tplt
var templateFS embed.FS

func LoadTemplate(tagName string) (*template.Template, error) {
	templateFilename := fmt.Sprintf("%s.tplt", tagName)

	data, err := templateFS.ReadFile(templateFilename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// No template
			return nil, fmt.Errorf("no template found for %s", tagName)
		}
		return nil, err
	}

	t := template.New(tagName).Funcs(template.FuncMap{
		"hasOption": hasOption,
		"unPointer": unPointer,
	})

	return t.Parse(string(data))
}

func hasOption(options []string, key string) bool {
	return slices.Contains(options, key)
}

func unPointer(ptrType string) string {
	return strings.TrimPrefix(ptrType, "*")
}

package siri

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type TemplateCache struct {
	root string
}

func NewTemplateCache(templatePath string) *TemplateCache {
	return &TemplateCache{root: templatePath}
}

type Data struct {
	Now       time.Time
	ClientRef string
}

var funcs = template.FuncMap{
	"dateTime": func(now time.Time) string {
		return now.Format(time.RFC3339)
	},
}

func (tc TemplateCache) ExecuteTemplate(name string, data Data) string {
	// TODO: improve error handling
	templateFile := filepath.Join(tc.root, name)
	templateName, err := filepath.Rel(tc.root, templateFile)
	if err != nil {
		return err.Error()
	}
	content, err := os.ReadFile(templateFile) //nolint gosec
	if err != nil {
		return err.Error()
	}
	t, err := template.New(templateName).Funcs(funcs).Parse(string(content))

	if err != nil {
		return err.Error()
	}

	var bytesBuffer bytes.Buffer
	terr := t.ExecuteTemplate(&bytesBuffer, name, data)
	if terr != nil {
		return terr.Error()
	}
	return bytesBuffer.String()
}

func (tc TemplateCache) TemplateNames() []string {
	root := filepath.Clean(tc.root)

	var templateNames []string
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error { //nolint errcheck I would ignore it anyway
		if d.Type().IsRegular() {
			if strings.HasSuffix(d.Name(), ".xml") {
				f, _ := filepath.Rel(root, path)
				templateNames = append(templateNames, f)
			}
		}
		return nil
	})

	return templateNames
}

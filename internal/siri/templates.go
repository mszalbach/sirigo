package siri

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type TemplateCache struct {
	root string
}

func NewTemplateCache(templatePath string) TemplateCache {
	return TemplateCache{root: templatePath}
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

func (tc TemplateCache) ExecuteTemplate(name string, data Data) (string, error) {
	templateFile := filepath.Join(tc.root, name)
	content, err := os.ReadFile(templateFile) //nolint gosec
	if err != nil {
		slog.Error("Could not read template file", slog.String("file", templateFile), slog.Any("error", err))
		return "", err
	}
	t, err := template.New(name).Funcs(funcs).Parse(string(content))

	if err != nil {
		slog.Error("Could not create template", slog.String("name", name), slog.Any("error", err))
		return "", err
	}

	var bytesBuffer bytes.Buffer
	terr := t.ExecuteTemplate(&bytesBuffer, name, data)
	if terr != nil {
		slog.Error("Could not execute template", slog.String("name", name), slog.Any("error", terr))
		return "", err
	}
	return bytesBuffer.String(), nil
}

func (tc TemplateCache) TemplateNames() ([]string, error) {
	root := filepath.Clean(tc.root)

	var templateNames []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error { //nolint errcheck I would ignore it anyway
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			if strings.HasSuffix(d.Name(), ".xml") {
				f, err := filepath.Rel(root, path)
				if err != nil {
					return err
				}
				templateNames = append(templateNames, f)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return templateNames, nil
}

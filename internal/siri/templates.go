package siri

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// TemplateCache used to execute file system templates for SIRI communication
type TemplateCache struct {
	root string
}

// NewTemplateCache creates new TemplateCache
func NewTemplateCache(templatePath string) TemplateCache {
	return TemplateCache{root: templatePath}
}

// Data used to render the templates
type Data struct {
	Now       time.Time
	ClientRef string
}

var funcs = template.FuncMap{
	"dateTime": func(now time.Time) string {
		return now.Format(time.RFC3339)
	},
	"addTime": func(now time.Time, duration string) time.Time {
		dur, err := time.ParseDuration(duration)
		if err != nil {
			return now
		}
		return now.Add(dur)
	},
}

// ExecuteTemplate finds the template and executes it with the data provided
func (tc TemplateCache) ExecuteTemplate(name string, data Data) (string, error) {
	templateFile := filepath.Join(tc.root, name)
	content, err := os.ReadFile(templateFile) //nolint gosec
	if err != nil {
		return "", fmt.Errorf("could not read template file %s: %w", templateFile, err)
	}
	t, err := template.New(name).Funcs(funcs).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("could not create template %s: %w", name, err)
	}

	var bytesBuffer bytes.Buffer
	terr := t.ExecuteTemplate(&bytesBuffer, name, data)
	if terr != nil {
		return "", fmt.Errorf("could not execute template %s: %w", name, err)
	}
	return bytesBuffer.String(), nil
}

// TemplateNames returns all found templates names from the root folder
func (tc TemplateCache) TemplateNames() ([]string, error) {
	root := filepath.Clean(tc.root)

	var templateNames []string
	err := filepath.WalkDir(
		root,
		func(path string, d os.DirEntry, err error) error {
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
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not get template names from %s: %w", root, err)
	}

	return templateNames, nil
}

// GetURLPathFromTemplate finds a comment with an url path
// used to specify where a SIRI client request should be sent to
func GetURLPathFromTemplate(template string) string {
	// Look for <!-- path: /siri/et.xml -->
	re := regexp.MustCompile(`(?s)<!--\s*path:\s*(.*?)\s*-->`)
	matches := re.FindStringSubmatch(template)
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

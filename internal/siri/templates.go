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

// TemplateCache is used to execute file-system templates for SIRI communication
type TemplateCache struct {
	root *os.Root
}

// NewTemplateCache creates a new TemplateCache
func NewTemplateCache(templatePath string) (TemplateCache, error) {
	root, err := os.OpenRoot(templatePath)
	if err != nil {
		return TemplateCache{}, err
	}
	return TemplateCache{root: root}, nil
}

// data is used to render the templates
type data struct {
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

// GetTemplate returns the content of a template on the filesystem
func (tc TemplateCache) GetTemplate(name string) (string, error) {
	content, err := tc.root.ReadFile(name) //nolint gosec
	if err != nil {
		return "", err
	}
	return string(content), err
}

// executeTemplate finds the template and executes it with the provided data
func executeTemplate(content string, data data) (string, error) {
	siriTemplate, err := template.New("siri").Funcs(funcs).Parse(content)
	if err != nil {
		return "", err
	}

	var bytesBuffer bytes.Buffer
	if err := siriTemplate.Execute(&bytesBuffer, data); err != nil {
		return "", err
	}
	return bytesBuffer.String(), nil
}

// TemplateNames returns all found template names from the root folder
func (tc TemplateCache) TemplateNames() ([]string, error) {
	var templateNames []string
	root := tc.root.Name()
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

var urlPathRegexp = regexp.MustCompile(`<!--\s*path:\s*(.*?)\s*-->`)

// GetURLPathFromTemplate finds a comment with an url path
// used to specify where a SIRI client request should be sent to
func GetURLPathFromTemplate(siriTemplate string) string {
	matches := urlPathRegexp.FindStringSubmatch(siriTemplate)
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

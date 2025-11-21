package siri_test

import (
	"testing"
	"time"

	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/stretchr/testify/assert"
)

var now = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func Test_returns_rendered_template_with_replaced_variables(t *testing.T) {
	// Given
	tc := siri.NewTemplateCache("testdata")

	// When
	actual := tc.ExecuteTemplate("siri/test.xml", siri.Data{Now: now, ClientRef: "testClient"})

	// Then
	expected := `<Siri>
  <time>2024-06-01T12:00:00Z</time>
  <client>testClient</client>
</Siri>`
	assert.Equal(t, expected, actual)
}

func Test_returns_rendered_template_when_empty_data_is_used(t *testing.T) {
	// Given
	tc := siri.NewTemplateCache("testdata")

	// When
	actual := tc.ExecuteTemplate("siri/test.xml", siri.Data{})

	// Then
	expected := `<Siri>
  <time>0001-01-01T00:00:00Z</time>
  <client></client>
</Siri>`
	assert.Equal(t, expected, actual)
}

func Test_returns_some_string_when_there_are_no_templates(t *testing.T) {
	// Given
	tc := siri.NewTemplateCache("testdata/empty")

	// When
	actual := tc.ExecuteTemplate("DOES-NOT-EXIST.xml", siri.Data{Now: now, ClientRef: "testClient"})

	// Then
	expected := "open testdata/empty/DOES-NOT-EXIST.xml: no such file or directory"
	assert.Equal(t, expected, actual)
}

func Test_returns_template_names(t *testing.T) {
	testCases := []struct {
		templatePath      string
		expectedTemplates []string
	}{
		{"testdata", []string{"siri/test.xml", "siri/test2.xml", "vdv453/ans/test.xml", "vdv453/test.xml"}},
		{"testdata/vdv453", []string{"ans/test.xml", "test.xml"}},
		{"testdata/empty", nil},
	}
	for _, tc := range testCases {
		t.Run(tc.templatePath, func(t *testing.T) {
			// Given
			cache := siri.NewTemplateCache(tc.templatePath)

			// When
			actual := cache.TemplateNames()

			// Then
			//TODO better examples in the testdata folder
			assert.Equal(t, tc.expectedTemplates, actual)
		})
	}
}

// Run with go test ./internal/siri  -fuzz=Fuzz
func Fuzz_template_cache(f *testing.F) {

	f.Add("testdata", "siri/test.xml")
	f.Fuzz(func(t *testing.T, templatePath string, templateName string) {
		tc := siri.NewTemplateCache(templatePath)
		success := assert.NotNil(t, tc)
		if success {
			tc.TemplateNames()
		}
		assert.NotEmpty(t, tc.ExecuteTemplate(templateName, siri.Data{}))
	})

}

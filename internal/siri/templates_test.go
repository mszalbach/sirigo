package siri

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func Test_returns_rendered_template_with_replaced_variables(t *testing.T) {
	// Given
	tc := NewTemplateCache("testdata")

	// When
	actual, err := tc.ExecuteTemplate("siri/test.xml", Data{Now: now, ClientRef: "testClient"})
	require.NoError(t, err)

	// Then
	expected := `<Siri>
  <time>2024-06-01T12:00:00Z</time>
  <futureTime>2024-06-01T12:05:00Z<futureTime>
  <client>testClient</client>
</Siri>`
	assert.Equal(t, expected, actual)
}

func Test_returns_rendered_template_when_empty_data_is_used(t *testing.T) {
	// Given
	tc := NewTemplateCache("testdata")

	// When
	actual, err := tc.ExecuteTemplate("siri/test.xml", Data{})
	require.NoError(t, err)

	// Then
	expected := `<Siri>
  <time>0001-01-01T00:00:00Z</time>
  <futureTime>0001-01-01T00:05:00Z<futureTime>
  <client></client>
</Siri>`
	assert.Equal(t, expected, actual)
}

func Test_returns_error_when_there_are_no_templates(t *testing.T) {
	// Given
	tc := NewTemplateCache("testdata/empty")

	// When
	_, err := tc.ExecuteTemplate("DOES-NOT-EXIST.xml", Data{Now: now, ClientRef: "testClient"})

	// Then
	assert.Error(t, err)
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
			cache := NewTemplateCache(tc.templatePath)

			// When
			actual, err := cache.TemplateNames()
			require.NoError(t, err)

			// Then
			// TODO better examples in the testdata folder
			assert.Equal(t, tc.expectedTemplates, actual)
		})
	}
}

func Test_can_extract_url_paths_from_strings(t *testing.T) {
	testCases := []struct {
		name            string
		template        string
		expectedURLPath string
	}{
		{"Empty string", "", ""},
		{"Only the url path comment", "<!-- path: /siri/et.xml -->", "/siri/et.xml"},
		{"Returns the first URL path found", "<!-- path: /siri/et.xml --><!-- path: /siri/vm.xml -->", "/siri/et.xml"},
		{"realistic XML example", `<!-- path: /siri/ca.xml -->
<Siri xmlns="http://www.siri.org.uk/siri" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" version="2.1">
	<SubscriptionRequest>
	</SubscriptionRequest>
</Siri>`, "/siri/ca.xml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualURLPath := GetURLPathFromTemplate(tc.template)
			assert.Equal(t, tc.expectedURLPath, actualURLPath)
		})
	}
}

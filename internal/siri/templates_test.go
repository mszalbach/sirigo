package siri

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func Test_returns_rendered_template_with_replaced_variables(t *testing.T) {
	// Given
	template := `<Siri>
  <time>{{ dateTime .Now }}</time>
  <futureTime>{{ dateTime (addTime .Now "5m") }}<futureTime>
  <client>{{ .ClientRef }}</client>
</Siri>`
	data := data{Now: now, ClientRef: "testClient"}

	// When
	actual, err := executeTemplate(template, data)
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
	template := `<Siri>
  <time>{{ dateTime .Now }}</time>
  <futureTime>{{ dateTime (addTime .Now "5m") }}<futureTime>
  <client>{{ .ClientRef }}</client>
</Siri>`
	data := data{}

	// When
	actual, err := executeTemplate(template, data)
	require.NoError(t, err)

	// Then
	expected := `<Siri>
  <time>0001-01-01T00:00:00Z</time>
  <futureTime>0001-01-01T00:05:00Z<futureTime>
  <client></client>
</Siri>`
	assert.Equal(t, expected, actual)
}

func Test_returns_empty_when_there_is_no_templates(t *testing.T) {
	// Given
	template := ""
	data := data{Now: now, ClientRef: "testClient"}

	// When
	actual, err := executeTemplate(template, data)
	require.NoError(t, err)

	// Then
	assert.Empty(t, actual)
}

func Test_returns_template_names(t *testing.T) {
	testCases := map[string]struct {
		templatePath      string
		expectedTemplates []string
	}{
		"root testdata": {
			"testdata",
			[]string{"siri/test.xml", "siri/test2.xml", "vdv453/ans/test.xml", "vdv453/test.xml"},
		},
		"using one subfolder": {"testdata/vdv453", []string{"ans/test.xml", "test.xml"}},
		"empty folder":        {"testdata/empty", nil},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Given
			cache, err := NewTemplateCache(tc.templatePath)
			require.NoError(t, err)

			// When
			actual, err := cache.TemplateNames()
			require.NoError(t, err)

			// Then
			assert.Equal(t, tc.expectedTemplates, actual)
		})
	}
}

func Test_can_extract_url_paths_from_strings(t *testing.T) {
	testCases := map[string]struct {
		template        string
		expectedURLPath string
	}{
		"Empty string":                     {"", ""},
		"Only the url path comment":        {"<!-- path: /siri/et.xml -->", "/siri/et.xml"},
		"Returns the first URL path found": {"<!-- path: /siri/et.xml --><!-- path: /siri/vm.xml -->", "/siri/et.xml"},
		"realistic XML example": {`<!-- path: /siri/ca.xml -->
<Siri xmlns="http://www.siri.org.uk/siri" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" version="2.1">
	<SubscriptionRequest>
	</SubscriptionRequest>
</Siri>`, "/siri/ca.xml"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actualURLPath := GetURLPathFromTemplate(tc.template)
			assert.Equal(t, tc.expectedURLPath, actualURLPath)
		})
	}
}

func Test_returns_content_of_file_below_root_folder(t *testing.T) {
	// Given
	cache, err := NewTemplateCache("testdata")
	require.NoError(t, err)

	// When
	actualContent, err := cache.GetTemplate("siri/test.xml")
	require.NoError(t, err)

	// Then
	expectedContent := `<Siri>
  <time>{{ dateTime .Now }}</time>
  <futureTime>{{ dateTime (addTime .Now "5m") }}<futureTime>
  <client>{{ .ClientRef }}</client>
</Siri>`
	assert.Equal(t, expectedContent, actualContent)
}

func Test_returns_error_when_file_does_not_exist(t *testing.T) {
	// Given
	cache, err := NewTemplateCache("testdata")
	require.NoError(t, err)

	// When
	_, terr := cache.GetTemplate("siri/DOES-NOT_EXIST.xml")

	// Then
	require.Error(t, terr)
}

func Test_should_not_leave_the_template_root_folder(t *testing.T) {
	// Given
	cache, err := NewTemplateCache("testdata")
	require.NoError(t, err)

	// ensure file outside of testdata exists
	_, terr := os.Stat("testdata/../../../README.md")
	require.NoError(t, terr)

	// When
	c, err := cache.GetTemplate("../../../README.md")

	assert.Empty(t, c)
	// Then
	require.Error(t, err)
}

package siri

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var now = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func Test_returns_rendered_template_with_replaced_variables(t *testing.T) {
	// Given
	tc := NewTemplateCache("testdata")

	// When
	actual := tc.ExecuteTemplate("siri/test.xml", Data{Now: now, ClientRef: "testClient"})

	// Then
	expected := `<Siri>
  <time>2024-06-01T12:00:00Z</time>
  <client>testClient</client>
</Siri>`
	assert.Equal(t, expected, actual)
}

func Test_returns_rendered_template_when_empty_data_is_used(t *testing.T) {
	// Given
	tc := NewTemplateCache("testdata")

	// When
	actual := tc.ExecuteTemplate("siri/test.xml", Data{})

	// Then
	expected := `<Siri>
  <time>0001-01-01T00:00:00Z</time>
  <client></client>
</Siri>`
	assert.Equal(t, expected, actual)
}

func Test_returns_some_string_when_there_are_no_templates(t *testing.T) {
	// Given
	tc := NewTemplateCache("testdata/empty")

	// When
	actual := tc.ExecuteTemplate("DOES-NOT-EXIST.xml", Data{Now: now, ClientRef: "testClient"})

	// Then
	assert.NotEmpty(t, actual)
}

func Test_returns_template_names(t *testing.T) {
	// Given
	tc := NewTemplateCache("testdata")

	// When
	actual := tc.TemplateNames()

	// Then
	//TODO better examples in the testdata folder
	expected := []string{"siri/test.xml", "siri/test2.xml", "vdv453/ans/test.xml", "vdv453/test.xml"}
	assert.Equal(t, expected, actual)
}

package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	row := map[string]string{"_raw": "foo", "index": "fooindex", "host": "barhost"}

	// Try to call Exec first, should error
	_, err := Exec("test", row)
	assert.EqualError(t, err, "Exec called for template 'test' but not found in cache")

	// Create a new test template
	err = New("test", "{{ ._raw }}")
	assert.NoError(t, err)
	temp, err := Exec("test", row)
	assert.NoError(t, err)
	assert.Equal(t, "foo", temp)

	// More complicated
	err = New("test2", "index={{ .index}} host={{ .host }} _raw={{ ._raw }}")
	assert.NoError(t, err)
	temp, err = Exec("test2", row)
	assert.NoError(t, err)
	assert.Equal(t, "index=fooindex host=barhost _raw=foo", temp)

	// JSON
	err = New("test3", "{{ json . | printf \"%s\" }}")
	assert.NoError(t, err)
	temp, err = Exec("test3", row)
	assert.NoError(t, err)
	assert.Equal(t, `{"_raw":"foo","host":"barhost","index":"fooindex"}`, temp)

	// Multiple variables, one replacement
	err = New("test4", "{{ ._raw }}{{ .foo }}")
	assert.NoError(t, err)
	temp, err = Exec("test4", row)
	assert.NoError(t, err)
	assert.Equal(t, "foo<no value>", temp)

	err = New("testheader", `{{ keys . | join "," }}`)
	assert.NoError(t, err)
	temp, err = Exec("testheader", row)
	assert.NoError(t, err)
	assert.Equal(t, "_raw,host,index", temp)
	// fmt.Println(temp)

	err = New("testvalues", `{{ values . | join "," }}`)
	assert.NoError(t, err)
	temp, err = Exec("testvalues", row)
	assert.NoError(t, err)
	assert.Equal(t, "foo,barhost,fooindex", temp)
	// fmt.Println(temp)

	// Test splunkhec template
	row = map[string]string{"_raw": "test raw", "_time": "1234567890.123", "host": "testhost", "source": "testsource"}
	err = New("splunkhec", "{{ splunkhec . }}")
	assert.NoError(t, err)
	temp, err = Exec("splunkhec", row)
	assert.NoError(t, err)
	assert.Equal(t, `{"event":"test raw","host":"testhost","source":"testsource","time":"1234567890.123"}`, temp)
}

func TestExists(t *testing.T) {
	// Test non-existent template
	exists := Exists("nonexistent")
	assert.False(t, exists, "Template 'nonexistent' should not exist")

	// Create a new template and verify it exists
	err := New("testexists", "test template")
	assert.NoError(t, err)
	exists = Exists("testexists")
	assert.True(t, exists, "Template 'testexists' should exist")
}

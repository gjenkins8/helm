/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"text/template"
)

func TestReadValues(t *testing.T) {
	doc := `# Test YAML parse
poet: "Coleridge"
title: "Rime of the Ancient Mariner"
stanza:
  - "at"
  - "length"
  - "did"
  - cross
  - an
  - Albatross

mariner:
  with: "crossbow"
  shot: "ALBATROSS"

water:
  water:
    where: "everywhere"
    nor: "any drop to drink"
`

	data, err := ReadValues([]byte(doc))
	require.NoError(t, err, "Error parsing bytes")
	matchValues(t, data)

	tests := []string{`poet: "Coleridge"`, "# Just a comment", ""}

	for _, tt := range tests {
		data, err = ReadValues([]byte(tt))
		if err != nil {
			require.NoError(t, err, "Error parsing bytes (%s)", tt)
		}
		if data == nil {
			assert.NotNil(t, m, `YAML string "%s" gave a nil map`, tt)
		}
	}
}

func TestReadValuesFile(t *testing.T) {
	data, err := ReadValuesFile("./testdata/coleridge.yaml")
	require.NoError(t, err, "Error reading YAML file")
	matchValues(t, data)
}

func ExampleValues() {
	doc := `
title: "Moby Dick"
chapter:
  one:
    title: "Loomings"
  two:
    title: "The Carpet-Bag"
  three:
    title: "The Spouter Inn"
`
	d, err := ReadValues([]byte(doc))
	if err != nil {
		panic(err)
	}
	ch1, err := d.Table("chapter.one")
	if err != nil {
		panic("could not find chapter one")
	}
	fmt.Print(ch1["title"])
	// Output:
	// Loomings
}

func TestTable(t *testing.T) {
	doc := `
title: "Moby Dick"
chapter:
  one:
    title: "Loomings"
  two:
    title: "The Carpet-Bag"
  three:
    title: "The Spouter Inn"
`
	d, err := ReadValues([]byte(doc))
	require.NoError(t, err, "Failed to parse the White Whale")

	if _, err := d.Table("title"); err == nil {
		require.Fail(t, "Title is not a table.")
	}

	if _, err := d.Table("chapter"); err != nil {
		require.NoError(t, err, "Failed to get the chapter table")
	}

	if v, err := d.Table("chapter.one"); err != nil {
		assert.NoError(t, err, "Failed to get chapter.one")
	} else if v["title"] != "Loomings" {
		assert.Equal(t, "Loomings", v["title"])
	}

	if _, err := d.Table("chapter.three"); err != nil {
		assert.NoError(t, err, "Chapter three is missing")
	}

	if _, err := d.Table("chapter.OneHundredThirtySix"); err == nil {
		assert.Fail(t, "I think you mean 'Epilogue'")
	}
}

func matchValues(t *testing.T, data map[string]any) {
	t.Helper()
	if data["poet"] != "Coleridge" {
		assert.Equal(t, "Coleridge", data["poet"])
	}

	if o, err := ttpl("{{len .stanza}}", data); err != nil {
		assert.NoError(t, err, "len stanza")
	} else if o != "6" {
		assert.Equal(t, int64(6), o)
	}

	if o, err := ttpl("{{.mariner.shot}}", data); err != nil {
		assert.NoError(t, err, ".mariner.shot")
	} else if o != "ALBATROSS" {
		assert.Fail(t, "Expected that mariner shot ALBATROSS")
	}

	if o, err := ttpl("{{.water.water.where}}", data); err != nil {
		assert.NoError(t, err, ".water.water.where: %s")
	} else if o != "everywhere" {
		assert.Fail(t, "Expected water water everywhere")
	}
}

func ttpl(tpl string, v map[string]any) (string, error) {
	var b bytes.Buffer
	tt := template.Must(template.New("t").Parse(tpl))
	err := tt.Execute(&b, v)
	return b.String(), err
}

func TestPathValue(t *testing.T) {
	doc := `
title: "Moby Dick"
chapter:
  one:
    title: "Loomings"
  two:
    title: "The Carpet-Bag"
  three:
    title: "The Spouter Inn"
`
	d, err := ReadValues([]byte(doc))
	require.NoError(t, err, "Failed to parse the White Whale")

	if v, err := d.PathValue("chapter.one.title"); err != nil {
		assert.NoError(t, err, "Got error instead of title")
	} else if v != "Loomings" {
		assert.Fail(t, "got wrong value for title")
	}
	if _, err := d.PathValue("chapter.one.doesnotexist"); err == nil {
		assert.Error(t, err, "Non-existent key should return error")
	}
	if _, err := d.PathValue("chapter.doesnotexist.one"); err == nil {
		assert.Error(t, err, "Non-existent key in middle of path should return error")
	}
	if _, err := d.PathValue(""); err == nil {
		assert.Fail(t, "Asking for the value from an empty path should yield an error")
	}
	if v, err := d.PathValue("title"); err == nil {
		if v != "Moby Dick" {
			assert.Fail(t, "Failed to return values for root key title")
		}
	}
}

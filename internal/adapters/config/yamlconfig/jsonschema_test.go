// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package yamlconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

// configSchemaPath is the published JSON Schema for .ci-parity.yml.
var configSchemaPath = filepath.Join("..", "..", "..", "..", "schemas", "config.schema.json")

// compareKeys reports where the keys the parser accepts (s) differ from the
// properties the JSON Schema (js) declares, walking both trees from path.
func compareKeys(t *testing.T, path string, s *schema, js map[string]any) {
	t.Helper()
	switch s.kind {
	case yaml.MappingNode:
		props, _ := js["properties"].(map[string]any)
		var want, got []string
		for k := range s.fields {
			want = append(want, k)
		}
		for k := range props {
			got = append(got, k)
		}
		slices.Sort(want)
		slices.Sort(got)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("%s: keys (-parser +schema):\n%s", describe(path), diff)
			return
		}
		for k, sub := range s.fields {
			child, _ := props[k].(map[string]any)
			compareKeys(t, join(path, k), sub, child)
		}
	case yaml.SequenceNode:
		items, _ := js["items"].(map[string]any)
		compareKeys(t, path+"[]", s.elem, items)
	}
}

func TestJSONSchemaMatchesTheKeysTheParserAccepts(t *testing.T) {
	// Arrange
	data, err := os.ReadFile(configSchemaPath)
	if err != nil {
		t.Fatal(err)
	}

	// Act
	var js map[string]any
	if err := json.Unmarshal(data, &js); err != nil {
		t.Fatalf("parse %s: %v", configSchemaPath, err)
	}

	// Assert
	compareKeys(t, "", fileSchema, js)
}

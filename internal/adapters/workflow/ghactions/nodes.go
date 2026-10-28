// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package ghactions

import (
	"gopkg.in/yaml.v3"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// get returns the value for key in mapping n, or nil.
func get(n *yaml.Node, key string) *yaml.Node {
	_, v := getPair(n, key)
	return v
}

// getPair returns the key and value nodes for key in mapping n, or nils.
func getPair(n *yaml.Node, key string) (*yaml.Node, *yaml.Node) {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil, nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i], n.Content[i+1]
		}
	}
	return nil, nil
}

// scalar returns the value of a scalar node, or "" for anything else.
func scalar(n *yaml.Node) string {
	if n == nil || n.Kind != yaml.ScalarNode {
		return ""
	}
	return n.Value
}

// inputs returns the scalar entries of the with: mapping n.
func inputs(n *yaml.Node) map[string]domain.Input {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}
	m := make(map[string]domain.Input, len(n.Content)/2)
	for i := 0; i+1 < len(n.Content); i += 2 {
		if v := n.Content[i+1]; v.Kind == yaml.ScalarNode {
			m[n.Content[i].Value] = domain.Input{Value: v.Value, Line: v.Line, Col: v.Column}
		}
	}
	return m
}

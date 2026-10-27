package yamlconfig

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// schema describes the expected shape of a config node so errors can name the
// offending key instead of a Go type.
type schema struct {
	kind   yaml.Kind
	fields map[string]*schema
	elem   *schema
}

var (
	scalar     = &schema{kind: yaml.ScalarNode}
	scalarList = &schema{kind: yaml.SequenceNode, elem: scalar}
)

var fileSchema = &schema{kind: yaml.MappingNode, fields: map[string]*schema{
	"runner":    scalar,
	"taskfile":  scalar,
	"workflows": scalarList,
	"invokers": {kind: yaml.MappingNode, fields: map[string]*schema{
		"shell": scalarList,
		"actions": {kind: yaml.SequenceNode, elem: &schema{kind: yaml.MappingNode, fields: map[string]*schema{
			"uses":  scalar,
			"input": scalar,
		}}},
	}},
	"tools": {kind: yaml.MappingNode, fields: map[string]*schema{
		"extend":  scalarList,
		"remove":  scalarList,
		"replace": scalarList,
	}},
	"transparent": scalarList,
	"allow": {kind: yaml.SequenceNode, elem: &schema{kind: yaml.MappingNode, fields: map[string]*schema{
		"match":     scalar,
		"reason":    scalar,
		"workflows": scalarList,
	}}},
	"baseline": scalar,
}}

// check validates n against s. Keys with a null value are removed from n so
// that decoding leaves their defaults in place.
func (s *schema) check(n *yaml.Node, path string) error {
	if n.Kind != s.kind {
		return fmt.Errorf("line %d: %s must be %s", n.Line, describe(path), kindName(s.kind))
	}
	switch s.kind {
	case yaml.MappingNode:
		kept := n.Content[:0]
		for i := 0; i+1 < len(n.Content); i += 2 {
			key, val := n.Content[i], n.Content[i+1]
			name := join(path, key.Value)
			sub, ok := s.fields[key.Value]
			if !ok {
				return fmt.Errorf("line %d: unknown key %q", key.Line, name)
			}
			if isNull(val) {
				continue
			}
			if err := sub.check(val, name); err != nil {
				return err
			}
			kept = append(kept, key, val)
		}
		n.Content = kept
	case yaml.SequenceNode:
		for i, item := range n.Content {
			if err := s.elem.check(item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func isNull(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.Tag == "!!null"
}

func join(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

func describe(path string) string {
	if path == "" {
		return "the config"
	}
	return path
}

func kindName(k yaml.Kind) string {
	switch k {
	case yaml.MappingNode:
		return "a mapping"
	case yaml.SequenceNode:
		return "a list"
	default:
		return "a string"
	}
}

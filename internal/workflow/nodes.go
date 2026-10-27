package workflow

import "gopkg.in/yaml.v3"

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

// stringMap returns the scalar entries of mapping n.
func stringMap(n *yaml.Node) map[string]string {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}
	m := make(map[string]string, len(n.Content)/2)
	for i := 0; i+1 < len(n.Content); i += 2 {
		if v := n.Content[i+1]; v.Kind == yaml.ScalarNode {
			m[n.Content[i].Value] = v.Value
		}
	}
	return m
}

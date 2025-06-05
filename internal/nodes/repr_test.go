package nodes

import (
	"strings"
	"testing"
)

func TestLeafValuesWithBrackets(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected string
	}{
		{
			name: "simple node",
			node: &Node{
				Key:   "test",
				Value: "value",
			},
			expected: "value",
		},
		{
			name: "node with children",
			node: &Node{
				Key: "parent",
				Children: []*Node{
					{Key: "child1", Value: "value1"},
					{Key: "child2", Value: "value2"},
				},
			},
			expected: "{child1: value1 child2: value2}",
		},
		{
			name: "nested children",
			node: &Node{
				Key: "root",
				Children: []*Node{
					{
						Key: "level1",
						Children: []*Node{
							{Key: "level2", Value: "value"},
						},
					},
				},
			},
			expected: "{level1: {level2: value}}",
		},
		{
			name: "empty children",
			node: &Node{
				Key:      "empty",
				Children: []*Node{},
			},
			expected: "",
		},
		{
			name: "long string truncation",
			node: &Node{
				Key:   "long",
				Value: strings.Repeat("a", MaxStringLength+10),
			},
			expected: strings.Repeat("a", MaxStringLength) + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LeafValuesWithBrackets(tt.node)
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLeafValuesOnly(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected string
	}{
		{
			name: "simple leaf node",
			node: &Node{
				Key:   "test",
				Value: "value",
			},
			expected: "value",
		},
		{
			name: "node with leaf children",
			node: &Node{
				Key: "parent",
				Children: []*Node{
					{Key: "child1", Value: "value1"},
					{Key: "child2", Value: "value2"},
				},
			},
			expected: "value1 value2",
		},
		{
			name: "mixed leaf and non-leaf children",
			node: &Node{
				Key: "root",
				Children: []*Node{
					{
						Key: "nonleaf",
						Children: []*Node{
							{Key: "nested", Value: "value"},
						},
					},
					{Key: "leaf", Value: "leafvalue"},
				},
			},
			expected: "value leafvalue",
		},
		{
			name: "no leaf children",
			node: &Node{
				Key: "root",
				Children: []*Node{
					{
						Key: "level1",
						Children: []*Node{
							{Key: "level2", Children: []*Node{}},
						},
					},
				},
			},
			expected: "",
		},
		{
			name: "long string truncation",
			node: &Node{
				Key:   "long",
				Value: strings.Repeat("a", MaxStringLength+10),
			},
			expected: strings.Repeat("a", MaxStringLength) + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LeafValuesOnly(tt.node)
			if result != tt.expected {
				t.Errorf("ShortString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestKeyCountOnly(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected string
	}{
		{
			name: "simple object with children",
			node: &Node{
				Key: "user",
				Children: []*Node{
					{Key: "name", Value: "John"},
					{Key: "age", Value: "30"},
				},
			},
			expected: MetadataPrefix + "(2 keys)",
		},
		{
			name: "array with items",
			node: &Node{
				Key: "items",
				Children: []*Node{
					{Key: "0", Value: "first"},
					{Key: "1", Value: "second"},
					{Key: "2", Value: "third"},
				},
			},
			expected: MetadataPrefix + "(3 items)",
		},
		{
			name: "node without children",
			node: &Node{
				Key:   "simple",
				Value: "value",
			},
			expected: "",
		},
		{
			name: "empty children array",
			node: &Node{
				Key:      "empty",
				Children: []*Node{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := KeyCountOnly(tt.node)
			if result != tt.expected {
				t.Errorf("KeyCountOnly() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestKeyNamesWithTypes(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected string
	}{
		{
			name: "object with mixed types",
			node: &Node{
				Key: "user",
				Children: []*Node{
					{Key: "name", Value: "John"},
					{Key: "age", Value: "30"},
					{Key: "active", Value: "true"},
				},
			},
			expected: MetadataPrefix + "(name:string, age:integer, active:boolean)",
		},
		{
			name: "array of strings",
			node: &Node{
				Key: "colors",
				Children: []*Node{
					{Key: "0", Value: "red"},
					{Key: "1", Value: "blue"},
					{Key: "2", Value: "green"},
				},
			},
			expected: MetadataPrefix + "(array of strings)",
		},
		{
			name: "array of numbers",
			node: &Node{
				Key: "scores",
				Children: []*Node{
					{Key: "0", Value: "95"},
					{Key: "1", Value: "87"},
					{Key: "2", Value: "92"},
				},
			},
			expected: MetadataPrefix + "(array of numbers)",
		},
		{
			name: "array of mixed types",
			node: &Node{
				Key: "mixed",
				Children: []*Node{
					{Key: "0", Value: "text"},
					{Key: "1", Value: "42"},
					{Key: "2", Value: "true"},
				},
			},
			expected: MetadataPrefix + "(array)",
		},
		{
			name: "object with nested objects",
			node: &Node{
				Key: "data",
				Children: []*Node{
					{Key: "config", Children: []*Node{{Key: "debug", Value: "false"}}},
					{Key: "items", Children: []*Node{{Key: "0", Value: "item1"}}},
				},
			},
			expected: MetadataPrefix + "(config:object, items:array)",
		},
		{
			name: "node without children",
			node: &Node{
				Key:   "simple",
				Value: "value",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := KeyNamesWithTypes(tt.node)
			if result != tt.expected {
				t.Errorf("KeyNamesWithTypes() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestKeyCountAndTypes(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected string
	}{
		{
			name: "object with mixed types",
			node: &Node{
				Key: "user",
				Children: []*Node{
					{Key: "name", Value: "John"},
					{Key: "age", Value: "30"},
				},
			},
			expected: MetadataPrefix + "(2 keys: name:string, age:integer)",
		},
		{
			name: "array of strings",
			node: &Node{
				Key: "tags",
				Children: []*Node{
					{Key: "0", Value: "work"},
					{Key: "1", Value: "personal"},
				},
			},
			expected: MetadataPrefix + "(2 items: array of strings)",
		},
		{
			name: "single item array",
			node: &Node{
				Key: "single",
				Children: []*Node{
					{Key: "0", Value: "42"},
				},
			},
			expected: MetadataPrefix + "(1 items: array of numbers)",
		},
		{
			name: "node without children",
			node: &Node{
				Key:   "simple",
				Value: "value",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := KeyCountAndTypes(tt.node)
			if result != tt.expected {
				t.Errorf("KeyCountAndTypes() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDirectChildrenKeys(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected string
	}{
		{
			name: "object with children",
			node: &Node{
				Key: "user",
				Children: []*Node{
					{Key: "name", Value: "John"},
					{Key: "age", Value: "30"},
					{Key: "active", Value: "true"},
				},
			},
			expected: "name age active",
		},
		{
			name: "array with items",
			node: &Node{
				Key: "items",
				Children: []*Node{
					{Key: "0", Value: "first"},
					{Key: "1", Value: "second"},
					{Key: "2", Value: "third"},
				},
			},
			expected: "0 1 2",
		},
		{
			name: "single child",
			node: &Node{
				Key: "parent",
				Children: []*Node{
					{Key: "only_child", Value: "value"},
				},
			},
			expected: "only_child",
		},
		{
			name: "node without children",
			node: &Node{
				Key:   "simple",
				Value: "value",
			},
			expected: "{}",
		},
		{
			name: "empty children array",
			node: &Node{
				Key:      "empty",
				Children: []*Node{},
			},
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DirectChildrenKeys(tt.node)
			if result != tt.expected {
				t.Errorf("DirectChildrenKeys() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetJSONType(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected string
	}{
		{
			name:     "string value",
			node:     &Node{Key: "test", Value: "hello"},
			expected: "string",
		},
		{
			name:     "integer value",
			node:     &Node{Key: "test", Value: "42"},
			expected: "integer",
		},
		{
			name:     "float value",
			node:     &Node{Key: "test", Value: "3.14"},
			expected: "number",
		},
		{
			name:     "boolean true",
			node:     &Node{Key: "test", Value: "true"},
			expected: "boolean",
		},
		{
			name:     "boolean false",
			node:     &Node{Key: "test", Value: "false"},
			expected: "boolean",
		},
		{
			name:     "null value",
			node:     &Node{Key: "test", Value: ""},
			expected: "null",
		},
		{
			name: "object with children",
			node: &Node{
				Key:      "test",
				Children: []*Node{{Key: "child", Value: "value"}},
			},
			expected: "object",
		},
		{
			name: "array with numeric keys",
			node: &Node{
				Key: "test",
				Children: []*Node{
					{Key: "0", Value: "first"},
					{Key: "1", Value: "second"},
				},
			},
			expected: "array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getJSONType(tt.node)
			if result != tt.expected {
				t.Errorf("getJSONType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsArray(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected bool
	}{
		{
			name: "numeric keys array",
			node: &Node{
				Children: []*Node{
					{Key: "0", Value: "first"},
					{Key: "1", Value: "second"},
					{Key: "2", Value: "third"},
				},
			},
			expected: true,
		},
		{
			name: "mixed keys not array",
			node: &Node{
				Children: []*Node{
					{Key: "name", Value: "John"},
					{Key: "age", Value: "30"},
				},
			},
			expected: false,
		},
		{
			name: "non-sequential numeric keys still array",
			node: &Node{
				Children: []*Node{
					{Key: "0", Value: "first"},
					{Key: "2", Value: "third"},
				},
			},
			expected: true,
		},
		{
			name: "empty children not array",
			node: &Node{
				Children: []*Node{},
			},
			expected: false,
		},
		{
			name: "no children not array",
			node: &Node{
				Value: "simple",
			},
			expected: false,
		},
		{
			name: "mixed numeric and string keys not array",
			node: &Node{
				Children: []*Node{
					{Key: "0", Value: "first"},
					{Key: "name", Value: "test"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isArray(tt.node)
			if result != tt.expected {
				t.Errorf("isArray() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetArrayElementTypes(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected string
	}{
		{
			name: "array of strings",
			node: &Node{
				Children: []*Node{
					{Key: "0", Value: "hello"},
					{Key: "1", Value: "world"},
				},
			},
			expected: "array of strings",
		},
		{
			name: "array of integers",
			node: &Node{
				Children: []*Node{
					{Key: "0", Value: "1"},
					{Key: "1", Value: "2"},
					{Key: "2", Value: "3"},
				},
			},
			expected: "array of numbers",
		},
		{
			name: "array of floats",
			node: &Node{
				Children: []*Node{
					{Key: "0", Value: "1.5"},
					{Key: "1", Value: "2.7"},
				},
			},
			expected: "array of numbers",
		},
		{
			name: "mixed type array",
			node: &Node{
				Children: []*Node{
					{Key: "0", Value: "hello"},
					{Key: "1", Value: "42"},
					{Key: "2", Value: "true"},
				},
			},
			expected: "array",
		},
		{
			name: "not an array",
			node: &Node{
				Children: []*Node{
					{Key: "name", Value: "John"},
					{Key: "age", Value: "30"},
				},
			},
			expected: "array",
		},
		{
			name: "empty array",
			node: &Node{
				Children: []*Node{},
			},
			expected: "array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getArrayElementTypes(tt.node)
			if result != tt.expected {
				t.Errorf("getArrayElementTypes() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetRepr(t *testing.T) {
	tests := []struct {
		name     string
		reprType string
		expected string // We'll test the function name by checking if it's not nil
	}{
		{"leaf-with-brackets", LeafValuesWithBracketsRepr, "LeafValuesWithBrackets"},
		{"leaf-only", LeafValuesOnlyRepr, "LeafValuesOnly"},
		{"children-keys", DirectChildrenKeysRepr, "DirectChildrenKeys"},
		{"key-count-only", KeyCountOnlyRepr, "KeyCountOnly"},
		{"key-names-with-types", KeyNamesWithTypesRepr, "KeyNamesWithTypes"},
		{"key-count-and-types", KeyCountAndTypesRepr, "KeyCountAndTypes"},
		{"unknown", "unknown-type", "LeafValuesOnly"}, // default case
	}

	// Create a test node
	testNode := &Node{
		Key: "test",
		Children: []*Node{
			{Key: "child", Value: "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reprFunc := GetRepr(tt.reprType)
			if reprFunc == nil {
				t.Errorf("GetRepr(%v) returned nil", tt.reprType)
				return
			}

			// Test that the function actually works
			result := reprFunc(testNode)
			if result == "" {
				t.Errorf("GetRepr(%v) returned empty string", tt.reprType)
			}
		})
	}
}

func TestTruncateIfNeeded(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short string",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "exactly max length",
			input:    strings.Repeat("a", MaxStringLength),
			expected: strings.Repeat("a", MaxStringLength),
		},
		{
			name:     "over max length",
			input:    strings.Repeat("a", MaxStringLength+10),
			expected: strings.Repeat("a", MaxStringLength) + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateIfNeeded(tt.input)
			if result != tt.expected {
				t.Errorf("truncateIfNeeded() = %v, want %v", result, tt.expected)
			}
		})
	}
}

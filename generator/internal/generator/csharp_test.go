package generator

import (
	"strings"
	"testing"
	"v1m-SCUG/internal/cache"
	"v1m-SCUG/internal/parser"
)

func TestGenerateCSharp(t *testing.T) {
	node := &parser.Node{
		Name:      "Root",
		Sanitized: "Root",
		Components: []cache.ComponentInfo{
			{ClassName: "MyComponent", IsPublic: true},
		},
		Children: []*parser.Node{
			{
				Name:      "Child",
				Sanitized: "Child",
				Components: []cache.ComponentInfo{
					{ClassName: "ChildComp", IsPublic: false},
				},
			},
		},
	}

	code := generateCSharp(node, "MyNamespace", "MyClass", "Prefabs/Root")

	expectedSubstrings := []string{
		"namespace MyNamespace",
		"public class MyClass",
		"public const string ResourcePath = \"Prefabs/Root\"",
		"private global::MyComponent _MyComponent",
		"public class Child_Obj",
		"internal global::ChildComp ChildComp => _ChildComp",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(code, s) {
			t.Errorf("generated code missing expected substring: %q", s)
		}
	}
}

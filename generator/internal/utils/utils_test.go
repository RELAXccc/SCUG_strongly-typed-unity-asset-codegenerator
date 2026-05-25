package utils

import "testing"

func TestSanitizeIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Tree", "Tree"},
		{"Tree (1)", "Tree_1"},
		{"1Tree", "_1Tree"},
		{"My Prefab", "My_Prefab"},
		{"", ""},
		{".", ""},
		{"System", "System_"},
		{"UnityEngine", "UnityEngine_"},
		{"Wrapper", "Wrapper_"},
	}

	for _, tc := range tests {
		got := SanitizeIdentifier(tc.input)
		if got != tc.expected {
			t.Errorf("SanitizeIdentifier(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

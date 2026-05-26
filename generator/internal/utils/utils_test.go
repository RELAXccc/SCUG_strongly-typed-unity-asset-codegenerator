package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetFastHash(t *testing.T) {
	// Test with non-existent file
	if hash := GetFastHash("non_existent_file.txt"); hash != "" {
		t.Errorf("GetFastHash for non-existent file should be empty, got %q", hash)
	}

	// Create a temporary file
	tmpfile, err := ioutil.TempFile("", "hash_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("hello world")); err != nil {
		t.Fatal(err)
	}

	hash1 := GetFastHash(tmpfile.Name())
	if hash1 == "" {
		t.Fatal("GetFastHash returned empty for an existing file")
	}

	// Change file and check if hash changes
	time.Sleep(10 * time.Millisecond) // ensure mod time changes
	if _, err := tmpfile.Write([]byte(" updated")); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	hash2 := GetFastHash(tmpfile.Name())
	if hash2 == hash1 {
		t.Errorf("GetFastHash did not change after file modification. Both hashes: %q", hash1)
	}
}

func TestGetSimpleName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"MyClass", "MyClass"},
		{"MyNamespace.MyClass", "MyClass"},
		{"Deep.Nested.Namespace.MyClass", "MyClass"},
		{"", ""},
	}

	for _, tc := range tests {
		if got := GetSimpleName(tc.input); got != tc.expected {
			t.Errorf("GetSimpleName(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestFindAssetsDir(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := ioutil.TempDir("", "find_assets_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	assetsDir := filepath.Join(tmpDir, "Assets")
	if err := os.Mkdir(assetsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Test 1: Inside the root directory
	os.Chdir(tmpDir)
	if got := FindAssetsDir(); got != "Assets" && got != filepath.Join(tmpDir, "Assets") {
		t.Errorf("FindAssetsDir from root = %q; want %q or %q", got, "Assets", filepath.Join(tmpDir, "Assets"))
	}

	// Test 2: Inside a sub-directory
	subDir := filepath.Join(tmpDir, "SubDir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	os.Chdir(subDir)
	if got := FindAssetsDir(); got != "../Assets" && got != filepath.Join(tmpDir, "Assets") {
		t.Errorf("FindAssetsDir from subdir = %q; want %q or %q", got, "../Assets", filepath.Join(tmpDir, "Assets"))
	}

	// Test 3: Inside a deeply nested directory
	deepDir := filepath.Join(subDir, "DeepDir")
	if err := os.Mkdir(deepDir, 0755); err != nil {
		t.Fatal(err)
	}
	os.Chdir(deepDir)
	if got := FindAssetsDir(); got == "" || filepath.Base(got) != "Assets" {
		t.Errorf("FindAssetsDir from deep dir = %q; expected something ending in 'Assets'", got)
	}

	// Reset to original working directory
	// In testing we don't necessarily know the original cwd, but assuming the root project dir
}

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

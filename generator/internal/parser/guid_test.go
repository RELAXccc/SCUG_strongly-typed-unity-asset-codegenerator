package parser

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"v1m-SCUG/internal/cache"
	"v1m-SCUG/internal/config"
)

func generateRandomHex(n int) string {
	var letters = []rune("0123456789abcdef")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestBuildGuidMap_Randomized(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	tmpDir, err := ioutil.TempDir("", "scug_guid_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	scriptsDir := filepath.Join(tmpDir, "Assets", "Scripts")
	resourcesDir := filepath.Join(tmpDir, "Assets", "Resources")
	os.MkdirAll(scriptsDir, 0755)
	os.MkdirAll(resourcesDir, 0755)

	c := cache.NewCache()
	cfg := &config.Config{
		ScanDirs:     []string{"Assets/Scripts"},
		ResourcesDir: "Assets/Resources",
		Workers:      2,
	}

	expectedMappings := make(map[string]string) // map GUID to Expected ClassName

	// 1. Generate random .cs and .cs.meta files
	for i := 0; i < 5; i++ {
		className := fmt.Sprintf("RandomClass%d", i)
		guid := generateRandomHex(32) // e.g. 83226d145b3fda14494bff1d5116f296
		expectedMappings[guid] = className

		csContent := fmt.Sprintf("public class %s {}", className)
		ioutil.WriteFile(filepath.Join(scriptsDir, className+".cs"), []byte(csContent), 0644)

		metaContent := fmt.Sprintf("fileFormatVersion: 2\nguid: %s\n", guid)
		ioutil.WriteFile(filepath.Join(scriptsDir, className+".cs.meta"), []byte(metaContent), 0644)
	}

	// 2. Generate random .prefab.meta files
	for i := 0; i < 5; i++ {
		prefabName := fmt.Sprintf("RandomPrefab%d", i)
		guid := generateRandomHex(32)
		// Prefabs in the root of Resources get their exact name
		expectedMappings[guid] = prefabName

		metaContent := fmt.Sprintf("fileFormatVersion: 2\nguid: %s\n", guid)
		ioutil.WriteFile(filepath.Join(resourcesDir, prefabName+".prefab.meta"), []byte(metaContent), 0644)
	}

	// Also test subdirectories in resources
	subDir := filepath.Join(resourcesDir, "UI", "Elements")
	os.MkdirAll(subDir, 0755)
	subGuid := generateRandomHex(32)
	expectedMappings[subGuid] = "UI.Elements.RandomNestedPrefab"

	metaContent := fmt.Sprintf("fileFormatVersion: 2\nguid: %s\n", subGuid)
	ioutil.WriteFile(filepath.Join(subDir, "RandomNestedPrefab.prefab.meta"), []byte(metaContent), 0644)

	// Run function
	BuildGuidMap(cfg, tmpDir, c)

	// Verify
	for guid, expectedClass := range expectedMappings {
		info, exists := c.GetMeta(guid)
		if !exists {
			t.Errorf("BuildGuidMap failed to extract GUID %s", guid)
			continue
		}
		if info.ClassName != expectedClass {
			t.Errorf("For GUID %s, expected class %s but got %s", guid, expectedClass, info.ClassName)
		}
		if !info.IsPublic {
			t.Errorf("Expected IsPublic to be true for generated class %s", expectedClass)
		}
	}
}

func TestGetPrefabClassName(t *testing.T) {
	tests := []struct {
		prefabPath   string
		resourcesDir string
		expected     string
	}{
		{"Assets/Resources/MyPrefab.prefab", "Assets/Resources", "MyPrefab"},
		{"Assets/Resources/UI/Menu.prefab", "Assets/Resources", "UI.Menu"},
		{"Assets/Resources/Environment/Trees/Pine_1.prefab", "Assets/Resources", "Environment.Trees.Pine_1"},
		{"/outside/path.prefab", "Assets/Resources", ""}, // invalid relative path
	}

	for _, tc := range tests {
		got := getPrefabClassName(tc.prefabPath, tc.resourcesDir)
		if got != tc.expected {
			t.Errorf("getPrefabClassName(%q, %q) = %q; want %q", tc.prefabPath, tc.resourcesDir, got, tc.expected)
		}
	}
}

func TestExtractClassInfo(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "scug_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name              string
		content           string
		filename          string
		expectedClassName string
		expectedPublic    bool
	}{
		{
			name: "Public class with namespace",
			content: `
namespace MyNamespace {
    public class MyClass : MonoBehaviour {}
}`,
			filename:          "MyClass.cs",
			expectedClassName: "MyNamespace.MyClass",
			expectedPublic:    true,
		},
		{
			name: "Internal class without namespace",
			content: `
class MyInternalClass {
}`,
			filename:          "MyInternalClass.cs",
			expectedClassName: "MyInternalClass",
			expectedPublic:    false,
		},
		{
			name: "Multiple classes, pick filename match",
			content: `
public class OtherClass {}
public class TargetClass {}
`,
			filename:          "TargetClass.cs",
			expectedClassName: "TargetClass",
			expectedPublic:    true,
		},
		{
			name: "Protected class (defaults to internal)",
			content: `
protected class ProtectedClass {}
`,
			filename:          "ProtectedClass.cs",
			expectedClassName: "ProtectedClass",
			expectedPublic:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tc.filename)
			if err := ioutil.WriteFile(path, []byte(tc.content), 0644); err != nil {
				t.Fatal(err)
			}

			className, isPublic := extractClassInfo(path)
			if className != tc.expectedClassName {
				t.Errorf("extractClassInfo() className = %q, want %q", className, tc.expectedClassName)
			}
			if isPublic != tc.expectedPublic {
				t.Errorf("extractClassInfo() isPublic = %v, want %v", isPublic, tc.expectedPublic)
			}
		})
	}
}

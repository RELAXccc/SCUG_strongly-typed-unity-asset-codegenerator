package generator

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"v1m-SCUG/internal/cache"
	"v1m-SCUG/internal/parser"
)

func TestGenerateTags(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test_tags")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	settingsDir := filepath.Join(tmpDir, "ProjectSettings")
	os.MkdirAll(settingsDir, 0755)
	
	outDir := filepath.Join(tmpDir, "Output")
	os.MkdirAll(outDir, 0755)

	tagContent := `%YAML 1.1
%TAG !u! tag:unity3d.com,2011:
--- !u!78 &1
TagManager:
  tags:
  - Player
  - Enemy
  - custom_tag
  layers:
`
	ioutil.WriteFile(filepath.Join(settingsDir, "TagManager.asset"), []byte(tagContent), 0644)

	GenerateTags(tmpDir, outDir)

	outFile := filepath.Join(outDir, "Global", "Tags.cs")
	content, err := ioutil.ReadFile(outFile)
	if err != nil {
		t.Fatalf("GenerateTags did not generate Tags.cs: %v", err)
	}

	code := string(content)
	expected := []string{
		"public static readonly Tags Player = new Tags(\"Player\");",
		"public static readonly Tags Enemy = new Tags(\"Enemy\");",
		"public static readonly Tags Custom_tag = new Tags(\"custom_tag\");",
	}

	for _, e := range expected {
		if !strings.Contains(code, e) {
			t.Errorf("GenerateTags output missing %q\nCode:\n%s", e, code)
		}
	}
}

func TestGenerateScenes(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test_scenes")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	settingsDir := filepath.Join(tmpDir, "ProjectSettings")
	os.MkdirAll(settingsDir, 0755)
	
	outDir := filepath.Join(tmpDir, "Output")
	os.MkdirAll(outDir, 0755)

	scenesContent := `%YAML 1.1
%TAG !u! tag:unity3d.com,2011:
--- !u!1041 &1
EditorBuildSettings:
  m_Scenes:
  - enabled: 1
    path: Assets/Scenes/MainMenu.unity
  - enabled: 0
    path: Assets/Scenes/Hidden.unity
  - enabled: 1
    path: Assets/Scenes/Level1.unity
`
	ioutil.WriteFile(filepath.Join(settingsDir, "EditorBuildSettings.asset"), []byte(scenesContent), 0644)

	GenerateScenes(tmpDir, outDir)

	outFile := filepath.Join(outDir, "Global", "Map.cs")
	content, err := ioutil.ReadFile(outFile)
	if err != nil {
		t.Fatalf("GenerateScenes did not generate Map.cs: %v", err)
	}

	code := string(content)
	expected := []string{
		"public static readonly Maps MainMenu = new Maps(\"MainMenu\");",
		"public static readonly Maps Level1 = new Maps(\"Level1\");",
	}
	
	notExpected := []string{
		"Hidden",
	}

	for _, e := range expected {
		if !strings.Contains(code, e) {
			t.Errorf("GenerateScenes output missing %q\nCode:\n%s", e, code)
		}
	}
	
	for _, e := range notExpected {
		if strings.Contains(code, e) {
			t.Errorf("GenerateScenes output should NOT contain %q", e)
		}
	}
}

func TestGenerateResources(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test_res")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	resDir := filepath.Join(tmpDir, "Resources")
	os.MkdirAll(filepath.Join(resDir, "Icons"), 0755)
	os.MkdirAll(filepath.Join(resDir, "Audio"), 0755)
	
	outDir := filepath.Join(tmpDir, "Output")
	os.MkdirAll(outDir, 0755)

	ioutil.WriteFile(filepath.Join(resDir, "Icons", "gold.png"), []byte(""), 0644)
	ioutil.WriteFile(filepath.Join(resDir, "Audio", "click.mp3"), []byte(""), 0644)
	ioutil.WriteFile(filepath.Join(resDir, "MyPrefab.prefab"), []byte(""), 0644)
	// Should ignore meta
	ioutil.WriteFile(filepath.Join(resDir, "MyPrefab.prefab.meta"), []byte(""), 0644)

	GenerateResources(resDir, outDir)

	outFile := filepath.Join(outDir, "Global", "Res.cs")
	content, err := ioutil.ReadFile(outFile)
	if err != nil {
		t.Fatalf("GenerateResources did not generate Res.cs: %v", err)
	}

	code := string(content)
	expected := []string{
		"public static class Icons",
		"public static Sprite Gold => Resources.Load<Sprite>(\"Icons/gold\");",
		"public static AudioClip Click => Resources.Load<AudioClip>(\"Audio/click\");",
		"public static GameObject MyPrefab => Resources.Load<GameObject>(\"MyPrefab\");",
	}
	
	if strings.Contains(code, "MyPrefab_meta") {
		t.Errorf("GenerateResources should ignore .meta files")
	}

	for _, e := range expected {
		if !strings.Contains(code, e) {
			t.Errorf("GenerateResources output missing %q\nCode:\n%s", e, code)
		}
	}
}

func TestGetTypeFromExt(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{".png", "Sprite"},
		{".jpg", "Sprite"},
		{".mat", "Material"},
		{".prefab", "GameObject"},
		{".mp3", "AudioClip"},
		{".wav", "AudioClip"},
		{".anim", "AnimationClip"},
		{".controller", "RuntimeAnimatorController"},
		{".asset", "ScriptableObject"},
		{".txt", "TextAsset"},
		{".json", "TextAsset"},
		{".ttf", "Font"},
		{".unknown", "Object"},
		{"", "Object"},
	}

	for _, tc := range tests {
		got := getTypeFromExt(tc.ext)
		if got != tc.expected {
			t.Errorf("getTypeFromExt(%q) = %q; want %q", tc.ext, got, tc.expected)
		}
	}
}

func TestProcessPrefabFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test_process_prefab")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	resDir := filepath.Join(tmpDir, "Resources")
	outDir := filepath.Join(tmpDir, "Output")
	os.MkdirAll(resDir, 0755)
	os.MkdirAll(outDir, 0755)

	prefabPath := filepath.Join(resDir, "UI", "MyTestPrefab.prefab")
	os.MkdirAll(filepath.Dir(prefabPath), 0755)

	prefabContent := `--- !u!1 &100000
GameObject:
  m_Name: MyTestPrefab
  m_Component:
  - component: {fileID: 100001}
--- !u!4 &100001
Transform:
  m_GameObject: {fileID: 100000}
`
	ioutil.WriteFile(prefabPath, []byte(prefabContent), 0644)

	c := cache.NewCache()

	// Run ProcessPrefabFile
	ProcessPrefabFile(prefabPath, resDir, outDir, c, false)

	// Check if file was generated
	outFilePath := filepath.Join(outDir, "UI", "MyTestPrefab.cs")
	if _, err := os.Stat(outFilePath); os.IsNotExist(err) {
		t.Errorf("ProcessPrefabFile did not generate %s", outFilePath)
	}

	// Run again with cache enabled to test cache skip logic
	ProcessPrefabFile(prefabPath, resDir, outDir, c, false)
}

func TestGenerateClass_ArrayGrouping(t *testing.T) {
	node := &parser.Node{
		Name:      "Root",
		Sanitized: "Root",
		Children: []*parser.Node{
			{Name: "Item0", Sanitized: "Item0"},
			{Name: "Item1", Sanitized: "Item1"},
			{Name: "Item2", Sanitized: "Item2"},
			{Name: "OtherChild", Sanitized: "OtherChild"},
		},
	}

	code := generateClass(node, "    ")

	expectedSubstrings := []string{
		"public object[] Item_Array => _Item_Array ??= new object[] { Item0, Item1, Item2 };",
		"public class Item0_Obj",
		"public class Item1_Obj",
		"public class OtherChild_Obj",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(code, s) {
			t.Errorf("generated code missing expected substring: %q\nCode:\n%s", s, code)
		}
	}
}

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

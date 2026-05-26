package parser

import (
	"io/ioutil"
	"os"
	"testing"
	"v1m-SCUG/internal/cache"
)

func TestParsePrefab(t *testing.T) {
	content := `--- !u!1 &100000
GameObject:
  m_Name: MyObject
--- !u!4 &100001
Transform:
  m_GameObject: {fileID: 100000}
--- !u!1001 &100002
PrefabInstance:
  m_Modification:
    m_TransformParent: {fileID: 0}
    m_Modifications:
    - target: {fileID: 12345}
      propertyPath: m_Name
      value: "OverriddenName"
      objectReference: {fileID: 0}
    - target: {fileID: 67890}
      propertyPath: m_IsActive
      value: 0
      objectReference: {fileID: 0}
`
	tmpfile, err := ioutil.TempFile("", "test_prefab")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	blocks, err := ParsePrefab(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParsePrefab failed: %v", err)
	}

	if len(blocks) != 3 {
		t.Errorf("expected 3 blocks, got %d", len(blocks))
	}

	if blocks[0].ClassID != "1" || blocks[0].FileID != "100000" {
		t.Errorf("block 0 mismatch: ClassID=%s, FileID=%s", blocks[0].ClassID, blocks[0].FileID)
	}

	if blocks[2].ClassID != "1001" || blocks[2].FileID != "100002" {
		t.Errorf("block 2 mismatch: ClassID=%s, FileID=%s", blocks[2].ClassID, blocks[2].FileID)
	}
}

func TestProcessBlocks(t *testing.T) {
	blocks := []Block{
		{
			ClassID: "1",
			FileID:  "10",
			Lines: []string{
				"  m_Name: RootObject",
				"  m_Component:",
				"  - component: {fileID: 20}",
			},
		},
		{
			ClassID: "4",
			FileID:  "20",
			Lines: []string{
				"  m_GameObject: {fileID: 10}",
				"  m_Father: {fileID: 0}",
				"  m_Children:",
				"  - {fileID: 21}",
			},
		},
		{
			ClassID: "1",
			FileID:  "11",
			Lines: []string{
				"  m_Name: ChildObject",
				"  m_Component:",
				"  - component: {fileID: 21}",
			},
		},
		{
			ClassID: "4",
			FileID:  "21",
			Lines: []string{
				"  m_GameObject: {fileID: 11}",
				"  m_Father: {fileID: 20}",
			},
		},
	}

	c := cache.NewCache()
	node := ProcessBlocks(blocks, c)

	if node == nil {
		t.Fatal("ProcessBlocks returned nil")
	}

	if node.Name != "RootObject" {
		t.Errorf("expected node name RootObject, got %s", node.Name)
	}

	if len(node.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(node.Children))
	}

	if node.Children[0].Name != "ChildObject" {
		t.Errorf("expected child node name ChildObject, got %s", node.Children[0].Name)
	}
}

func TestProcessBlocks_PrefabInstance(t *testing.T) {
	blocks := []Block{
		{
			ClassID: "1",
			FileID:  "10",
			Lines: []string{
				"  m_Name: RootObject",
				"  m_Component:",
				"  - component: {fileID: 20}",
			},
		},
		{
			ClassID: "4",
			FileID:  "20",
			Lines: []string{
				"  m_GameObject: {fileID: 10}",
				"  m_Father: {fileID: 0}",
				"  m_Children:",
				"  - {fileID: 21}",
			},
		},
		{
			ClassID: "4",
			FileID:  "21",
			Lines: []string{
				"  m_GameObject: {fileID: 0}", // Stripped transform
				"  m_Father: {fileID: 20}",
				"  m_CorrespondingSourceObject: {fileID: 1234, guid: abcdef123456, type: 3}",
				"  m_PrefabInstance: {fileID: 30}",
			},
		},
		{
			ClassID: "1001",
			FileID:  "30",
			Lines: []string{
				"  m_Modification:",
				"    m_Modifications:",
				"    - target: {fileID: 12345}",
				"      propertyPath: m_Name",
				"      value: OverriddenPrefabName",
			},
		},
	}

	c := cache.NewCache()
	c.MetaMapping["abcdef123456"] = cache.ComponentInfo{ClassName: "MyNestedPrefab", IsPublic: true}
	node := ProcessBlocks(blocks, c)

	if node == nil {
		t.Fatal("ProcessBlocks returned nil")
	}

	if len(node.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(node.Children))
	}
	
	child := node.Children[0]

	if child.Name != "OverriddenPrefabName" {
		t.Errorf("expected node name OverriddenPrefabName, got %s", child.Name)
	}
	if child.WrapperType != "MyNestedPrefab" {
		t.Errorf("expected wrapper type MyNestedPrefab, got %s", child.WrapperType)
	}
}

func TestProcessBlocks_MissingGuid(t *testing.T) {
	blocks := []Block{
		{
			ClassID: "1",
			FileID:  "10",
			Lines: []string{
				"  m_Name: RootObject",
				"  m_Component:",
				"  - component: {fileID: 20}",
			},
		},
		{
			ClassID: "4",
			FileID:  "20",
			Lines: []string{
				"  m_GameObject: {fileID: 10}",
				"  m_Father: {fileID: 0}",
				"  m_Children:",
				"  - {fileID: 21}",
			},
		},
		{
			ClassID: "4",
			FileID:  "21",
			Lines: []string{
				"  m_GameObject: {fileID: 0}", // Stripped transform
				"  m_Father: {fileID: 20}",
				"  m_CorrespondingSourceObject: {fileID: 1234, guid: deadbeef123456, type: 3}",
			},
		},
	}

	c := cache.NewCache()
	node := ProcessBlocks(blocks, c)

	if node == nil {
		t.Fatal("ProcessBlocks returned nil")
	}

	if len(node.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(node.Children))
	}
	
	child := node.Children[0]

	if child.Name != "NestedPrefab" {
		t.Errorf("expected fallback node name NestedPrefab, got %s", child.Name)
	}
	if child.WrapperType != "" {
		t.Errorf("expected empty wrapper type, got %s", child.WrapperType)
	}
}

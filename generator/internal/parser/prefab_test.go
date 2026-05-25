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

	if len(blocks) != 2 {
		t.Errorf("expected 2 blocks, got %d", len(blocks))
	}

	if blocks[0].ClassID != "1" || blocks[0].FileID != "100000" {
		t.Errorf("block 0 mismatch: ClassID=%s, FileID=%s", blocks[0].ClassID, blocks[0].FileID)
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
}

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRun_FullScan(t *testing.T) {
	// Use testdata as project context
	cwd, _ := os.Getwd()
	projectRoot, _ := filepath.Abs(filepath.Join(cwd, "..", "..", "..", "testdata"))
	
	// Change working directory to project root so utils.FindAssetsDir works
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to testdata directory: %v", err)
	}
	defer os.Chdir(cwd)

	// Clean up any existing output/cache for a fresh test
	os.Remove("scug_cache.json")
	
	args := []string{"scug"}
	exitCode := Run(args)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify some output exists
	if _, err := os.Stat("Assets/CustomOutput/Global/Map.cs"); os.IsNotExist(err) {
		t.Errorf("Map.cs was not generated in full scan")
	}
}

func TestRun_Targeted(t *testing.T) {
	// Use testdata as project context
	cwd, _ := os.Getwd()
	projectRoot, _ := filepath.Abs(filepath.Join(cwd, "..", "..", "..", "testdata"))
	
	// Change working directory to project root so utils.FindAssetsDir works
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to testdata directory: %v", err)
	}
	defer os.Chdir(cwd)

	// Targeted mode with a specific prefab
	args := []string{"scug", "Assets/Resources/Prefabs/EnvironmentPrefabs/Tree.prefab"}
	exitCode := Run(args)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRun_Targeted_NonExistent(t *testing.T) {
	// Use testdata as project context
	cwd, _ := os.Getwd()
	projectRoot, _ := filepath.Abs(filepath.Join(cwd, "..", "..", "..", "testdata"))
	
	// Change working directory to project root so utils.FindAssetsDir works
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to testdata directory: %v", err)
	}
	defer os.Chdir(cwd)

	// This should fail gracefully or at least not crash
	args := []string{"scug", "NonExistent.prefab"}
	exitCode := Run(args)
	
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for non-existent file (worker handles it), got %d", exitCode)
	}
}

func TestRun_NoAssets(t *testing.T) {
	// Run in a temp dir with no Assets folder
	tmpDir, _ := os.MkdirTemp("", "scug_no_assets")
	defer os.RemoveAll(tmpDir)

	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	exitCode := Run([]string{"scug"})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 when Assets dir is missing, got %d", exitCode)
	}
}

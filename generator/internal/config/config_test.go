package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "scug_config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "scug.json")

	// 1. Test missing config uses defaults
	cfg1 := LoadConfig(tmpDir)
	if cfg1.CacheFile != "scug_cache.json" {
		t.Errorf("Expected default CacheFile scug_cache.json, got %s", cfg1.CacheFile)
	}
	if cfg1.Workers <= 0 {
		t.Errorf("Expected workers > 0")
	}

	// 2. Write custom config
	customJSON := `{"cache_file": "custom_cache.json", "workers": 4, "output_dir": "CustomOut"}`
	ioutil.WriteFile(configPath, []byte(customJSON), 0644)

	cfg2 := LoadConfig(tmpDir)
	if cfg2.CacheFile != "custom_cache.json" {
		t.Errorf("Expected custom_cache.json, got %s", cfg2.CacheFile)
	}
	if cfg2.Workers != 4 {
		t.Errorf("Expected 4 workers, got %d", cfg2.Workers)
	}
	if cfg2.OutputDir != "CustomOut" {
		t.Errorf("Expected CustomOut, got %s", cfg2.OutputDir)
	}
}

func TestGetAbsolutePath(t *testing.T) {
	cfg := &Config{}
	
	// Unix relative
	abs1 := cfg.GetAbsolutePath("/home/user/project", "Assets/Scripts")
	if abs1 != filepath.Join("/home/user/project", "Assets/Scripts") {
		t.Errorf("Failed absolute resolution, got %s", abs1)
	}

	// Unix Absolute (should return as-is)
	abs2 := cfg.GetAbsolutePath("/home/user/project", "/absolute/path")
	if abs2 != "/absolute/path" {
		t.Errorf("Failed absolute passthrough, got %s", abs2)
	}
}

package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

// Config holds the configuration options for the SCUG generator.
type Config struct {
	DisableCache bool     `json:"disable_cache"`
	CacheFile    string   `json:"cache_file"`
	ResourcesDir string   `json:"resources_dir"`
	OutputDir    string   `json:"output_dir"`
	ScanDirs     []string `json:"scan_dirs"`
	Workers      int      `json:"workers"`
}

// LoadConfig reads the configuration from scug.json in the project root.
// If the file doesn't exist, it returns a default configuration.
func LoadConfig(projectRoot string) *Config {
	// Default configuration
	cfg := &Config{
		DisableCache: false,
		CacheFile:    "scug_cache.json",
		ResourcesDir: "Assets/Resources",
		OutputDir:    "Assets/Scripts/v2/UX/generated",
		ScanDirs: []string{
			"Assets/Scripts",
			"Library/PackageCache",
			"Packages",
		},
		Workers: runtime.NumCPU(),
	}

	configPath := filepath.Join(projectRoot, "scug.json")
	if data, err := ioutil.ReadFile(configPath); err == nil {
		json.Unmarshal(data, cfg)
	}

	// Validate and fallback config values
	if cfg.Workers <= 0 {
		cfg.Workers = runtime.NumCPU()
	}
	if cfg.CacheFile == "" {
		cfg.CacheFile = "scug_cache.json"
	}
	if cfg.ResourcesDir == "" {
		cfg.ResourcesDir = "Assets/Resources"
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "Assets/Scripts/v2/UX/generated"
	}
	if len(cfg.ScanDirs) == 0 {
		cfg.ScanDirs = []string{"Assets/Scripts", "Library/PackageCache", "Packages"}
	}

	return cfg
}

// GetAbsolutePath resolves a potentially relative path against the project root.
func (c *Config) GetAbsolutePath(projectRoot, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(projectRoot, path)
}

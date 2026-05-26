package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"v1m-SCUG/internal/cache"
	"v1m-SCUG/internal/config"
	"v1m-SCUG/internal/generator"
	"v1m-SCUG/internal/parser"
	"v1m-SCUG/internal/utils"
)

// SCUG (Super Cool Unity Generator) entry point.
// This tool parses Unity prefabs and generates strongly-typed C# wrappers.
func main() {
	start := time.Now()

	// Locate the Assets directory to establish project context.
	assetsDir := utils.FindAssetsDir()
	if assetsDir == "" {
		fmt.Println("Error: Could not find Assets directory.")
		os.Exit(1)
	}

	// Calculate project root based on Assets directory location.
	projectRoot := filepath.Dir(assetsDir)

	// Load configuration.
	cfg := config.LoadConfig(projectRoot)

	// Resolve cache path. If it's the default, keep it in the current working dir for backwards compatibility.
	cachePath := cfg.CacheFile
	if filepath.IsAbs(cachePath) {
		// already absolute
	} else if cachePath != "scug_cache.json" {
		cachePath = cfg.GetAbsolutePath(projectRoot, cachePath)
	}

	// Load persistent cache to speed up subsequent runs (if not disabled).
	var c *cache.Cache
	if !cfg.DisableCache {
		c = cache.LoadCache(cachePath)
	} else {
		c = cache.NewCache() // Use empty cache in memory if disabled
	}

	// Define key directories based on config.
	resourcesDir := cfg.GetAbsolutePath(projectRoot, cfg.ResourcesDir)
	outputDir := cfg.GetAbsolutePath(projectRoot, cfg.OutputDir)

	// Scan Assets for .cs.meta files to resolve Unity GUIDs to C# classes.
	parser.BuildGuidMap(cfg, projectRoot, c)

	// Initialize worker pool for parallel processing of prefabs.
	numWorkers := cfg.Workers
	var wg sync.WaitGroup
	var mu sync.Mutex
	count := 0

	// Handle Targeted Mode (triggered when specific files are passed as arguments).
	files := os.Args[1:]
	if len(files) > 0 {
		fmt.Printf("Processing %d targeted prefabs...\n", len(files))

		pathsChan := make(chan string, len(files))
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for rawPath := range pathsChan {
					if !strings.HasSuffix(strings.ToLower(rawPath), ".prefab") {
						continue
					}
					fullPath := filepath.Join(projectRoot, rawPath)
					if _, err := os.Stat(fullPath); os.IsNotExist(err) {
						fmt.Printf("Error: File does not exist: %s\n", rawPath)
						os.Exit(1)
					}
					generator.ProcessPrefabFile(fullPath, resourcesDir, outputDir, c, cfg.DisableCache)
					mu.Lock()
					count++
					mu.Unlock()
				}
			}()
		}

		for _, f := range files {
			pathsChan <- f
		}
		close(pathsChan)
		wg.Wait()

		if !cfg.DisableCache {
			c.Save(cachePath)
		}
		fmt.Printf("Done! Processed %d prefabs in %v.\n", count, time.Since(start))
	} else {
		generator.GenerateScenes(projectRoot, outputDir)
		generator.GenerateTags(projectRoot, outputDir)
		generator.GenerateResources(resourcesDir, outputDir)

		// Handle Full Scan Mode (triggered when no arguments are provided).
		fmt.Println("No arguments provided. Running full scan...")

		var prefabPaths []string
		filepath.Walk(resourcesDir, func(pathStr string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if strings.HasSuffix(pathStr, ".prefab") {
				prefabPaths = append(prefabPaths, pathStr)
			}
			return nil
		})

		pathsChan := make(chan string, len(prefabPaths))
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for pathStr := range pathsChan {
					generator.ProcessPrefabFile(pathStr, resourcesDir, outputDir, c, cfg.DisableCache)
					mu.Lock()
					count++
					mu.Unlock()
				}
			}()
		}

		for _, p := range prefabPaths {
			pathsChan <- p
		}
		close(pathsChan)
		wg.Wait()

		// Cleanup files in the output directory that no longer have a corresponding prefab.
		fmt.Println("Cleaning up stale files...")
		filepath.Walk(outputDir, func(pathStr string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if strings.HasSuffix(pathStr, ".cs") {
				if strings.Contains(filepath.ToSlash(pathStr), "/Global/") {
					return nil
				}
				slashPath := strings.ToLower(filepath.ToSlash(pathStr))
				if !c.IsFileGenerated(slashPath) {
					fmt.Println("Deleting stale file:", pathStr)
					os.Remove(pathStr)
				}
			}
			return nil
		})

		if !cfg.DisableCache {
			c.Save(cachePath)
		}
		fmt.Printf("Done! Full scan processed %d prefabs in %v.\n", count, time.Since(start))
	}
}

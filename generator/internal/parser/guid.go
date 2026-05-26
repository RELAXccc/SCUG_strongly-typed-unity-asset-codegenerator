package parser

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"v1m-SCUG/internal/cache"
	"v1m-SCUG/internal/config"
	"v1m-SCUG/internal/utils"
)

func BuildGuidMap(cfg *config.Config, projectRoot string, c *cache.Cache) {
	var scanDirs []string
	for _, dir := range cfg.ScanDirs {
		scanDirs = append(scanDirs, cfg.GetAbsolutePath(projectRoot, dir))
	}

	resourcesDir := cfg.GetAbsolutePath(projectRoot, cfg.ResourcesDir)

	fmt.Printf("Updating GUID Mapping... ")
	
	// Create a worker pool for meta file processing
	numWorkers := cfg.Workers
	pathsChan := make(chan string, 100)
	var wg sync.WaitGroup
	var updateCount int32

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pathStr := range pathsChan {
				info, err := os.Stat(pathStr)
				if err != nil {
					continue
				}
				modTime := info.ModTime().UnixNano()
				if c.IsMetaUpToDate(pathStr, modTime) {
					continue
				}

				content, err := ioutil.ReadFile(pathStr)
				if err == nil {
					guidMatch := regexp.MustCompile(`guid:\s*([a-f0-9]+)`).FindStringSubmatch(string(content))
					if len(guidMatch) > 1 {
						guid := guidMatch[1]
						
						if strings.HasSuffix(pathStr, ".cs.meta") {
							csPath := strings.TrimSuffix(pathStr, ".meta")
							className, isPublic := extractClassInfo(csPath)
							if className != "" {
								c.SetMeta(guid, cache.ComponentInfo{ClassName: className, IsPublic: isPublic}, pathStr, modTime)
								atomic.AddInt32(&updateCount, 1)
							}
						} else if strings.HasSuffix(pathStr, ".prefab.meta") {
							prefabPath := strings.TrimSuffix(pathStr, ".meta")
							className := getPrefabClassName(prefabPath, resourcesDir)
							if className != "" {
								c.SetMeta(guid, cache.ComponentInfo{ClassName: className, IsPublic: true}, pathStr, modTime)
								atomic.AddInt32(&updateCount, 1)
							}
						}
					}
				}
			}
		}()
	}

	for _, dir := range scanDirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}

		filepath.WalkDir(dir, func(pathStr string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if strings.HasSuffix(pathStr, ".cs.meta") {
				pathsChan <- pathStr
			}
			return nil
		})
	}

	// Scan Resources for .prefab.meta
	filepath.WalkDir(resourcesDir, func(pathStr string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(pathStr, ".prefab.meta") {
			pathsChan <- pathStr
		}
		return nil
	})

	close(pathsChan)
	wg.Wait()

	fmt.Printf("found %d new/changed items.\n", atomic.LoadInt32(&updateCount))
}

func getPrefabClassName(prefabPath, resourcesDir string) string {
	relPath, err := filepath.Rel(resourcesDir, prefabPath)
	if err != nil {
		return ""
	}
	slashPath := filepath.ToSlash(relPath)
	dirPart := path.Dir(slashPath)

	var nsParts []string
	if dirPart != "." {
		rawParts := strings.Split(dirPart, "/")
		for _, p := range rawParts {
			if s := utils.SanitizeIdentifier(p); s != "" {
				nsParts = append(nsParts, s)
			}
		}
	}
	className := utils.SanitizeIdentifier(strings.TrimSuffix(path.Base(slashPath), ".prefab"))

	fullClass := className
	if len(nsParts) > 0 {
		fullClass = strings.Join(nsParts, ".") + "." + className
	}
	return fullClass
}

func extractClassInfo(csPath string) (string, bool) {
	content, err := ioutil.ReadFile(csPath)
	if err != nil {
		return "", false
	}
	text := string(content)

	ns := ""
	nsMatch := regexp.MustCompile(`namespace\s+([a-zA-Z0-9_.]+)`).FindStringSubmatch(text)
	if len(nsMatch) > 1 {
		ns = nsMatch[1]
	}

	expectedClassName := strings.TrimSuffix(filepath.Base(csPath), ".cs")

	// Find all class declarations
	classMatches := regexp.MustCompile(`(public|internal|private|protected)?\s*class\s+([a-zA-Z0-9_]+)`).FindAllStringSubmatch(text, -1)

	if len(classMatches) > 0 {
		var bestMatch []string
		for _, match := range classMatches {
			if match[2] == expectedClassName {
				bestMatch = match
				break
			}
		}

		if bestMatch == nil {
			bestMatch = classMatches[0]
		}

		accessibility := strings.TrimSpace(bestMatch[1])
		className := bestMatch[2]

		// In C#, classes without a modifier are 'internal' by default.
		isPublic := (accessibility == "public")

		fullClass := className
		if ns != "" {
			fullClass = ns + "." + className
		}
		return fullClass, isPublic
	}
	return "", false
}

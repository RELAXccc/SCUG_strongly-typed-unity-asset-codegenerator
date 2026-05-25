package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func GetFastHash(pathStr string) string {
	info, err := os.Stat(pathStr)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d_%d", info.ModTime().UnixNano(), info.Size())
}

func SanitizeIdentifier(name string) string {
	if name == "" || name == "." {
		return ""
	}
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]+`)
	sanitized := reg.ReplaceAllString(name, "_")
	sanitized = strings.Trim(sanitized, "_")

	if len(sanitized) > 0 && sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "_" + sanitized
	}
	if len(sanitized) > 0 {
		sanitized = strings.ToUpper(sanitized[:1]) + sanitized[1:]
	}

	switch sanitized {
	case "System", "UnityEngine", "Wrapper", "Object", "Type", "Resources", "Namespace":
		sanitized += "_"
	}
	return sanitized
}

func GetSimpleName(fullName string) string {
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}

func FindAssetsDir() string {
	if _, err := os.Stat("Assets"); err == nil {
		return "Assets"
	}
	if _, err := os.Stat("../Assets"); err == nil {
		return "../Assets"
	}
	curr, _ := os.Getwd()
	for i := 0; i < 3; i++ {
		if _, err := os.Stat(filepath.Join(curr, "Assets")); err == nil {
			return filepath.Join(curr, "Assets")
		}
		curr = filepath.Dir(curr)
	}
	return ""
}

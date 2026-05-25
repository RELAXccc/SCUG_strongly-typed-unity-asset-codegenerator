package generator

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"v1m-SCUG/internal/cache"
	"v1m-SCUG/internal/parser"
	"v1m-SCUG/internal/utils"
)

// generateClass recursively generates C# nested classes for the GameObject hierarchy.
// It handles component properties, child GameObject wrappers, and child grouping.
func generateClass(node *parser.Node, indent string) string {
	var sb strings.Builder

	// Track component types to handle multiple components of the same type on one GameObject.
	compCounts := make(map[string]int)
	for _, comp := range node.Components {
		fullComp := "global::" + comp.ClassName
		simple := utils.GetSimpleName(comp.ClassName)
		propName := simple

		// Deduplicate components by adding a numerical suffix if needed.
		if count, exists := compCounts[simple]; exists {
			compCounts[simple] = count + 1
			propName = fmt.Sprintf("%s_%d", simple, count)
		} else {
			compCounts[simple] = 1
		}

		// Ensure component property names don't collide with child GameObject names.
		for _, child := range node.Children {
			if child.Sanitized == propName {
				propName += "Comp"
				break
			}
		}

		// Determine visibility based on the source script's accessibility.
		visibility := "public"
		if !comp.IsPublic {
			visibility = "internal"
		}

		// Generate a private backing field and a public/internal property with lazy initialization.
		sb.WriteString(fmt.Sprintf("%sprivate %s _%s;\n", indent, fullComp, propName))
		sb.WriteString(fmt.Sprintf("%s%s %s %s => _%s != null ? _%s : (_%s = transform?.GetComponent<%s>());\n", indent, visibility, fullComp, propName, propName, propName, propName, fullComp))
	}

	// Group children by base name for indexed arrays (e.g. Child0, Child1 -> Child_Array)
	// This allows for easier access to repeating elements in UI lists or grids.
	indexedGroups := make(map[string][]int)
	re := regexp.MustCompile(`^(.*?)(\d+)$`)
	for _, child := range node.Children {
		matches := re.FindStringSubmatch(child.Name)
		if len(matches) == 3 {
			baseName := matches[1]
			idx, _ := strconv.Atoi(matches[2])
			indexedGroups[baseName] = append(indexedGroups[baseName], idx)
		}
	}

	for base, indices := range indexedGroups {
		if len(indices) < 2 {
			delete(indexedGroups, base)
			continue
		}
		
		sanitizedBase := utils.SanitizeIdentifier(base)
		sort.Ints(indices)
		
		// Generate the array property that aggregates the individual child objects.
		sb.WriteString(fmt.Sprintf("\n%sprivate object[] _%s_Array;\n", indent, sanitizedBase))
		sb.WriteString(fmt.Sprintf("%spublic object[] %s_Array => _%s_Array ??= new object[] { ", indent, sanitizedBase, sanitizedBase))
		for i, idx := range indices {
			if i > 0 { sb.WriteString(", ") }
			sb.WriteString(fmt.Sprintf("%s%d", sanitizedBase, idx))
		}
		sb.WriteString(" };\n")
	}

	// Generate properties for each child GameObject.
	for _, child := range node.Children {
		if child.WrapperType != "" {
			// If the child is a nested prefab, we link to its own generated wrapper.
			sb.WriteString(fmt.Sprintf("\n%sprivate global::%s.Wrapper _%s;\n", indent, child.WrapperType, child.Sanitized))
			sb.WriteString(fmt.Sprintf("%spublic global::%s.Wrapper %s => _%s ??= global::%s.Get(transform?.Find(\"%s\")?.gameObject);\n", indent, child.WrapperType, child.Sanitized, child.Sanitized, child.WrapperType, child.Name))
		} else {
			// Otherwise, generate a nested class for this GameObject to maintain hierarchy.
			sb.WriteString(fmt.Sprintf("\n%spublic class %s_Obj\n%s{\n", indent, child.Sanitized, indent))
			sb.WriteString(fmt.Sprintf("%s    public global::UnityEngine.Transform transform { get; private set; }\n", indent))
			sb.WriteString(fmt.Sprintf("%s    public global::UnityEngine.GameObject gameObject => transform?.gameObject;\n", indent))
			sb.WriteString(fmt.Sprintf("%s    public %s_Obj(global::UnityEngine.Transform t) { transform = t; }\n", indent, child.Sanitized))
			sb.WriteString(fmt.Sprintf("%s    public void Destroy() { if (gameObject != null) global::UnityEngine.Object.Destroy(gameObject); }\n", indent))
			sb.WriteString(fmt.Sprintf("%s    public void SetActive(bool active) { if (gameObject != null) gameObject.SetActive(active); }\n", indent))
			sb.WriteString(fmt.Sprintf("%s    public T GetComponent<T>() where T : global::UnityEngine.Component => transform?.GetComponent<T>();\n", indent))
			sb.WriteString(fmt.Sprintf("%s    public Wrapper_Obj GetObj(string name) => new Wrapper_Obj(transform?.Find(name));\n\n", indent))

			// Recursively generate the hierarchy.
			sb.WriteString(generateClass(child, indent+"    "))
			sb.WriteString(fmt.Sprintf("%s}\n", indent))

			// Lazy-loaded property to access the child wrapper.
			sb.WriteString(fmt.Sprintf("\n%sprivate %s_Obj _%s;\n", indent, child.Sanitized, child.Sanitized))
			sb.WriteString(fmt.Sprintf("%spublic %s_Obj %s => _%s ??= new %s_Obj(transform?.Find(\"%s\"));\n", indent, child.Sanitized, child.Sanitized, child.Sanitized, child.Sanitized, child.Name))
		}
	}

	return sb.String()
}

// generateCSharp creates the full file content for the C# wrapper class.
func generateCSharp(node *parser.Node, namespace, className, resourcePath string) string {
	var sb strings.Builder

	sb.WriteString("// Auto-generated by SCUG\n")
	sb.WriteString("// DO NOT EDIT MANUALLY!\n\n")
	sb.WriteString("using System;\n")
	sb.WriteString("using UnityEngine;\n\n")

	// Wrap in namespace if provided.
	if namespace != "" && namespace != "Resources" {
		sb.WriteString(fmt.Sprintf("namespace %s\n{\n", namespace))
		sb.WriteString(fmt.Sprintf("    public class %s\n    {\n", className))
	} else {
		sb.WriteString(fmt.Sprintf("public class %s\n{\n", className))
	}

	indent := "        "
	if namespace == "" || namespace == "Resources" {
		indent = "    "
	}

	// Static helpers for loading and instantiating the prefab.
	sb.WriteString(fmt.Sprintf("%spublic const string ResourcePath = \"%s\";\n\n", indent, resourcePath))
	sb.WriteString(fmt.Sprintf("%spublic static global::UnityEngine.GameObject Load() => global::UnityEngine.Resources.Load<global::UnityEngine.GameObject>(ResourcePath);\n", indent))
	sb.WriteString(fmt.Sprintf("%spublic static global::UnityEngine.GameObject Instantiate(global::UnityEngine.Transform parent = null) { var prefab = Load(); return prefab != null ? global::UnityEngine.Object.Instantiate(prefab, parent) : null; }\n\n", indent))

	// Generic object wrapper for dynamic runtime access when static structure is insufficient.
	sb.WriteString(fmt.Sprintf("%spublic class Wrapper_Obj\n%s{\n", indent, indent))
	sb.WriteString(fmt.Sprintf("%s    public global::UnityEngine.Transform transform { get; private set; }\n", indent))
	sb.WriteString(fmt.Sprintf("%s    public global::UnityEngine.GameObject gameObject => transform?.gameObject;\n", indent))
	sb.WriteString(fmt.Sprintf("%s    public Wrapper_Obj(global::UnityEngine.Transform t) { transform = t; }\n", indent))
	sb.WriteString(fmt.Sprintf("%s    public void Destroy() { if (gameObject != null) global::UnityEngine.Object.Destroy(gameObject); }\n", indent))
	sb.WriteString(fmt.Sprintf("%s    public void SetActive(bool active) { if (gameObject != null) gameObject.SetActive(active); }\n", indent))
	sb.WriteString(fmt.Sprintf("%s    public T GetComponent<T>() where T : global::UnityEngine.Component => transform?.GetComponent<T>();\n", indent))
	sb.WriteString(fmt.Sprintf("%s    public Wrapper_Obj GetObj(string name) => new Wrapper_Obj(transform?.Find(name));\n", indent))
	sb.WriteString(fmt.Sprintf("%s}\n\n", indent))

	// The main Wrapper class that holds the root and starts the hierarchy.
	sb.WriteString(fmt.Sprintf("%spublic class Wrapper\n%s{\n", indent, indent))
	sb.WriteString(fmt.Sprintf("%s    public global::UnityEngine.GameObject Root { get; private set; }\n", indent))
	sb.WriteString(fmt.Sprintf("%s    public global::UnityEngine.Transform transform { get; private set; }\n", indent))
	sb.WriteString(fmt.Sprintf("%s    public global::UnityEngine.GameObject gameObject => transform?.gameObject;\n\n", indent))

	sb.WriteString(fmt.Sprintf("%s    public Wrapper(global::UnityEngine.GameObject root)\n%s    {\n", indent, indent))
	sb.WriteString(fmt.Sprintf("%s        Root = root;\n", indent))
	sb.WriteString(fmt.Sprintf("%s        transform = root != null ? root.transform : null;\n%s    }\n", indent, indent))
	sb.WriteString(fmt.Sprintf("%s    public void Destroy() { if (gameObject != null) global::UnityEngine.Object.Destroy(gameObject); }\n", indent))
	sb.WriteString(fmt.Sprintf("%s    public void SetActive(bool active) { if (gameObject != null) gameObject.SetActive(active); }\n", indent))
	sb.WriteString(fmt.Sprintf("%s    public T GetComponent<T>() where T : global::UnityEngine.Component => transform?.GetComponent<T>();\n", indent))
	sb.WriteString(fmt.Sprintf("%s    public Wrapper_Obj GetObj(string name) => new Wrapper_Obj(transform?.Find(name));\n\n", indent))

	// Generate all nested classes and properties.
	sb.WriteString(generateClass(node, indent+"    "))
	sb.WriteString(fmt.Sprintf("%s}\n\n", indent))

	// Factory methods for obtaining a wrapper instance.
	sb.WriteString(fmt.Sprintf("%spublic static Wrapper Get(global::UnityEngine.GameObject root) => new Wrapper(root);\n", indent))
	sb.WriteString(fmt.Sprintf("%spublic static Wrapper Get(global::UnityEngine.Component component) => new Wrapper(component?.gameObject);\n", indent))

	if namespace != "" && namespace != "Resources" {
		sb.WriteString("    }\n}\n")
	} else {
		sb.WriteString("}\n")
	}

	return sb.String()
}

func ProcessPrefabFile(prefabPath, resourcesDir, outputDir string, c *cache.Cache) {
	relPath, _ := filepath.Rel(resourcesDir, prefabPath)
	slashPath := filepath.ToSlash(relPath)
	resPath := strings.TrimSuffix(slashPath, ".prefab")

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
	ns := strings.Join(nsParts, ".")

	className := utils.SanitizeIdentifier(strings.TrimSuffix(path.Base(slashPath), ".prefab"))

	outDirPart := ""
	for _, p := range nsParts {
		outDirPart = filepath.Join(outDirPart, p)
	}
	outFilePath := filepath.Join(outputDir, outDirPart, className+".cs")

	// Normalize path for cache lookup
	normalizedOutPath := strings.ToLower(filepath.ToSlash(outFilePath))

	hash := utils.GetFastHash(prefabPath)
	if c.GetFileHash(prefabPath) == hash {
		if _, err := os.Stat(outFilePath); err == nil {
			c.MarkFileGenerated(normalizedOutPath)
			return
		}
	}

	blocks, err := parser.ParsePrefab(prefabPath)
	if err != nil || len(blocks) == 0 {
		return
	}

	rootNode := parser.ProcessBlocks(blocks, c)
	if rootNode == nil {
		return
	}

	csharpCode := generateCSharp(rootNode, ns, className, resPath)

	os.MkdirAll(filepath.Dir(outFilePath), 0755)
	ioutil.WriteFile(outFilePath, []byte(csharpCode), 0644)
	fmt.Println("Generated:", filepath.ToSlash(outFilePath))

	c.SetFileHash(prefabPath, hash)
	c.MarkFileGenerated(normalizedOutPath)
}

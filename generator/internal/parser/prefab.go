package parser

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"v1m-SCUG/internal/cache"
	"v1m-SCUG/internal/utils"
)

var classIDMap = map[string]string{
	"1":   "UnityEngine.GameObject",
	"4":   "UnityEngine.Transform",
	"8":   "UnityEngine.AudioSource",
	"20":  "UnityEngine.Camera",
	"54":  "UnityEngine.Rigidbody",
	"65":  "UnityEngine.BoxCollider",
	"212": "UnityEngine.SpriteRenderer",
	"222": "UnityEngine.CanvasRenderer",
	"223": "UnityEngine.Canvas",
	"224": "UnityEngine.RectTransform",
	"328": "UnityEngine.Video.VideoPlayer",
}

type Block struct {
	ClassID string
	FileID  string
	Lines   []string
}

// ParsePrefab reads a Unity .prefab file and splits it into individual YAML blocks.
// Unity YAML files use '---' as a separator followed by the object's FileID and ClassID.
// This function returns a slice of raw blocks for further processing.
func ParsePrefab(pathStr string) ([]Block, error) {
	content, err := ioutil.ReadFile(pathStr)
	if err != nil {
		return nil, err
	}

	var blocks []Block
	var currentBlock *Block

	lines := strings.Split(string(content), "\n")
	blockStartRe := regexp.MustCompile(`^--- !u!(\d+)\s+&(\d+)`)

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if matches := blockStartRe.FindStringSubmatch(line); len(matches) > 2 {
			if currentBlock != nil {
				blocks = append(blocks, *currentBlock)
			}
			currentBlock = &Block{ClassID: matches[1], FileID: matches[2]}
		} else if currentBlock != nil {
			currentBlock.Lines = append(currentBlock.Lines, line)
		}
	}
	if currentBlock != nil {
		blocks = append(blocks, *currentBlock)
	}
	return blocks, nil
}

type GameObject struct {
	FileID     string
	Name       string
	Components []string
}

type Transform struct {
	FileID         string
	GameObject     string
	Father         string
	Children       []string
	Guid           string
	PrefabInstance string
}

type PrefabInstance struct {
	FileID        string
	NameOverrides map[string]string
}

type Node struct {
	Name        string
	Sanitized   string
	Components  []cache.ComponentInfo
	Children    []*Node
	WrapperType string
}

func ProcessBlocks(blocks []Block, c *cache.Cache) *Node {
	gos := make(map[string]*GameObject)
	transforms := make(map[string]*Transform)
	components := make(map[string]cache.ComponentInfo)
	prefabs := make(map[string]*PrefabInstance)

	for _, b := range blocks {
		switch b.ClassID {
		case "1":
			goObj := &GameObject{FileID: b.FileID}
			inComp := false
			for _, line := range b.Lines {
				if strings.HasPrefix(line, "  m_Name:") {
					name := strings.TrimSpace(strings.TrimPrefix(line, "  m_Name:"))
					if strings.HasPrefix(name, `"`) && strings.HasSuffix(name, `"`) {
						name = name[1 : len(name)-1]
					} else if strings.HasPrefix(name, `'`) && strings.HasSuffix(name, `'`) {
						name = name[1 : len(name)-1]
					}
					goObj.Name = name
				} else if strings.HasPrefix(line, "  m_Component:") {
					inComp = true
				} else if inComp {
					if strings.HasPrefix(line, "  - component: {fileID:") {
						match := regexp.MustCompile(`fileID:\s*(-?\d+)`).FindStringSubmatch(line)
						if len(match) > 1 {
							goObj.Components = append(goObj.Components, match[1])
						}
					} else {
						inComp = false
					}
				}
			}
			gos[b.FileID] = goObj
		case "4", "224":
			tr := &Transform{FileID: b.FileID}
			inChild := false
			for _, line := range b.Lines {
				if strings.HasPrefix(line, "  m_GameObject:") {
					match := regexp.MustCompile(`fileID:\s*(-?\d+)`).FindStringSubmatch(line)
					if len(match) > 1 {
						tr.GameObject = match[1]
					}
				} else if strings.HasPrefix(line, "  m_Father:") {
					match := regexp.MustCompile(`fileID:\s*(-?\d+)`).FindStringSubmatch(line)
					if len(match) > 1 {
						tr.Father = match[1]
					}
				} else if strings.HasPrefix(line, "  m_CorrespondingSourceObject:") {
					match := regexp.MustCompile(`guid:\s*([a-f0-9]+)`).FindStringSubmatch(line)
					if len(match) > 1 {
						tr.Guid = match[1]
					}
				} else if strings.HasPrefix(line, "  m_PrefabInstance:") {
					match := regexp.MustCompile(`fileID:\s*(-?\d+)`).FindStringSubmatch(line)
					if len(match) > 1 {
						tr.PrefabInstance = match[1]
					}
				} else if strings.HasPrefix(line, "  m_Children:") {
					inChild = true
				} else if inChild {
					if strings.HasPrefix(line, "  - {fileID:") {
						match := regexp.MustCompile(`fileID:\s*(-?\d+)`).FindStringSubmatch(line)
						if len(match) > 1 {
							tr.Children = append(tr.Children, match[1])
						}
					} else {
						inChild = false
					}
				}
			}
			transforms[b.FileID] = tr
		case "1001":
			pi := &PrefabInstance{FileID: b.FileID, NameOverrides: make(map[string]string)}
			curTarget := ""
			isNamePath := false
			for _, line := range b.Lines {
				if strings.HasPrefix(line, "    - target: {fileID:") {
					match := regexp.MustCompile(`fileID:\s*(-?\d+)`).FindStringSubmatch(line)
					if len(match) > 1 {
						curTarget = match[1]
					}
					isNamePath = false
				} else if strings.HasPrefix(line, "      propertyPath: m_Name") {
					isNamePath = true
				} else if curTarget != "" && isNamePath && strings.HasPrefix(line, "      value:") {
					name := strings.TrimSpace(strings.TrimPrefix(line, "      value:"))
					if strings.HasPrefix(name, `"`) && strings.HasSuffix(name, `"`) {
						name = name[1 : len(name)-1]
					}
					pi.NameOverrides[curTarget] = name
					curTarget = ""
					isNamePath = false
				} else if strings.HasPrefix(line, "    - ") {
					// Reset on next modification block if not matched
					curTarget = ""
					isNamePath = false
				}
			}
			prefabs[b.FileID] = pi
		case "114":
			className := ""
			guid := ""
			for _, line := range b.Lines {
				if strings.HasPrefix(line, "  m_EditorClassIdentifier:") {
					className = strings.TrimSpace(strings.TrimPrefix(line, "  m_EditorClassIdentifier:"))
					if strings.Contains(className, "::") {
						parts := strings.Split(className, "::")
						className = parts[1]
					}
				} else if strings.HasPrefix(line, "  m_Script:") {
					match := regexp.MustCompile(`guid:\s*([a-f0-9]+)`).FindStringSubmatch(line)
					if len(match) > 1 {
						guid = match[1]
					}
				}
			}
			if guid != "" {
				if info, ok := c.MetaMapping[guid]; ok {
					components[b.FileID] = info
				} else if className != "" {
					components[b.FileID] = cache.ComponentInfo{ClassName: className, IsPublic: true}
				}
			} else if className != "" {
				components[b.FileID] = cache.ComponentInfo{ClassName: className, IsPublic: true}
			}
		default:
			if name, ok := classIDMap[b.ClassID]; ok {
				components[b.FileID] = cache.ComponentInfo{ClassName: name, IsPublic: true}
			}
		}
	}

	var rootTr *Transform
	for _, tr := range transforms {
		if tr.Father == "0" || tr.Father == "" {
			if _, ok := gos[tr.GameObject]; ok {
				rootTr = tr
				break
			}
		}
	}

	if rootTr == nil {
		return nil
	}

	return buildTree(rootTr, transforms, gos, components, prefabs, c)
}

func buildTree(tr *Transform, transforms map[string]*Transform, gos map[string]*GameObject, components map[string]cache.ComponentInfo, prefabs map[string]*PrefabInstance, c *cache.Cache) *Node {
	goObj, ok := gos[tr.GameObject]
	if !ok {
		// Stripped transform? Check if it has a guid
		if tr.Guid != "" {
			node := &Node{
				Name: "NestedPrefab",
			}
			if info, ok := c.MetaMapping[tr.Guid]; ok {
				node.Name = utils.GetSimpleName(info.ClassName)
				node.WrapperType = info.ClassName
			}

			// Check for name override in PrefabInstance
			if pi, ok := prefabs[tr.PrefabInstance]; ok {
				// We need to know which target in the prefab corresponds to the root
				// Often m_CorrespondingSourceObject guid is the prefab itself.
				// In Unity YAML, the target fileID in Pi corresponds to the fileID in the original prefab.
				// This is hard to map without reading the original prefab.
				// However, if there's only one name override or if it matches the base name, we can guess.
				// For now, let's look for any name override in this PrefabInstance.
				for _, name := range pi.NameOverrides {
					node.Name = name
					break
				}
			}
			node.Sanitized = utils.SanitizeIdentifier(node.Name)
			return node
		}
		return nil
	}

	node := &Node{
		Name:      goObj.Name,
		Sanitized: utils.SanitizeIdentifier(goObj.Name),
	}

	for _, compID := range goObj.Components {
		if info, exists := components[compID]; exists {
			if info.ClassName != "UnityEngine.Transform" && info.ClassName != "UnityEngine.RectTransform" {
				node.Components = append(node.Components, info)
			}
		}
	}

	childCounts := make(map[string]int)
	for _, childID := range tr.Children {
		if childTr, exists := transforms[childID]; exists {
			if childNode := buildTree(childTr, transforms, gos, components, prefabs, c); childNode != nil {
				name := childNode.Sanitized
				if count, ok := childCounts[name]; ok {
					childCounts[name] = count + 1
					childNode.Sanitized = fmt.Sprintf("%s_%d", name, count)
				} else {
					childCounts[name] = 1
				}
				node.Children = append(node.Children, childNode)
			}
		}
	}

	return node
}

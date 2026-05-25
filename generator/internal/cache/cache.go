package cache

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

type ComponentInfo struct {
	ClassName string `json:"class_name"`
	IsPublic  bool   `json:"is_public"`
}

type Cache struct {
	mu             sync.RWMutex
	MetaMapping    map[string]ComponentInfo `json:"meta_mapping"`
	FileHashes     map[string]string        `json:"file_hashes"`
	MetaModTimes   map[string]int64         `json:"meta_mod_times"`
	GeneratedFiles map[string]bool          `json:"-"`
}

func NewCache() *Cache {
	return &Cache{
		MetaMapping:    make(map[string]ComponentInfo),
		FileHashes:     make(map[string]string),
		MetaModTimes:   make(map[string]int64),
		GeneratedFiles: make(map[string]bool),
	}
}

func LoadCache(pathStr string) *Cache {
	data, err := ioutil.ReadFile(pathStr)
	if err != nil {
		return NewCache()
	}
	var c Cache
	if err := json.Unmarshal(data, &c); err != nil {
		return NewCache()
	}
	if c.MetaMapping == nil {
		c.MetaMapping = make(map[string]ComponentInfo)
	}
	if c.FileHashes == nil {
		c.FileHashes = make(map[string]string)
	}
	if c.MetaModTimes == nil {
		c.MetaModTimes = make(map[string]int64)
	}
	c.GeneratedFiles = make(map[string]bool)
	return &c
}

func (c *Cache) Save(pathStr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, _ := json.MarshalIndent(c, "", "  ")
	ioutil.WriteFile(pathStr, data, 0644)
}

func (c *Cache) GetMeta(guid string) (ComponentInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	info, ok := c.MetaMapping[guid]
	return info, ok
}

func (c *Cache) SetMeta(guid string, info ComponentInfo, modPath string, modTime int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.MetaMapping[guid] = info
	c.MetaModTimes[modPath] = modTime
}

func (c *Cache) IsMetaUpToDate(path string, modTime int64) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MetaModTimes[path] == modTime
}

func (c *Cache) GetFileHash(path string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.FileHashes[path]
}

func (c *Cache) SetFileHash(path string, hash string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FileHashes[path] = hash
}

func (c *Cache) MarkFileGenerated(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.GeneratedFiles[path] = true
}

func (c *Cache) IsFileGenerated(path string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GeneratedFiles[path]
}

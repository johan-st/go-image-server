package images

import (
	"fmt"
	"time"
	//github.com/hashicorp/golang-lru/v2
	// TODO: consider using hashicorp lru or ARC for cache
	// build my own cache will be a good exercise. If I match the interface of hashicorp lru,
	// I can easily switch to it later. Or even let the user of this package give me a lru
	// eqvivalent on creation.
)

// ImageHandler will try to keep the cache within these limits but does not guarantee it.
//
// note: Use 0 (zero) to explicitly set default.
// TODO: default cache rules are never clearing any cache
type CacheRules struct {
	MaxTimeSinceUse   time.Duration // Max age in seconds		(default: 0, unlimited)
	MaxTotalCacheSize Size          // Max cache size in bytes	(default: 0 unlimited)
	MaxNum            int           // Max number of images	(default: 0, unlimited)
}

type cache struct {
	size            Size
	numberOfObjects int
	cap             int
	objects         []cacheObject
}

func (c cache) String() string {
	return fmt.Sprintf(`cache:
	totalSize       %s
	numberOfObjects %d
	cap             %d
	objects         %d
`, c.size, c.numberOfObjects, c.cap, len(c.objects))
}

type cacheObject struct {
	path         string
	size         Size
	lastAccessed time.Time
}

func (co cacheObject) String() string {
	return fmt.Sprintf(`cacheObject 
	path         %s
	size         %s
	lastAccessed %s
`,
		co.path,
		co.size,
		co.lastAccessed)
}

func newCache(capacity int) cache {
	return cache{
		size:            0,
		numberOfObjects: 0,
		cap:             capacity,
		objects:         make([]cacheObject, 0, capacity),
	}
}

func (c *cache) add(co cacheObject) {
	c.objects = append(c.objects, co)
	c.size += co.size
	c.numberOfObjects++
}

func (c *cache) get(path string) (cacheObject, error) {
	for i, o := range c.objects {
		if o.path == "" {
			continue
		}
		if o.path == path {
			c.objects[i].lastAccessed = time.Time{}
			return o, nil
		}
	}
	return cacheObject{}, fmt.Errorf("cache could not locate object with path %s", path)
}

func (c *cache) del(path string) {
	for i, co := range c.objects {
		if co.path == path {
			c.objects[i].path = ""
			return
		}
	}
}

// TODO: clean uo this api. Somewhat confusing to return an empty cacheObject?
func (c *cache) delLRU() cacheObject {
	if c.size == 0 {
		return cacheObject{}
	}

	lru := time.Now()
	lruI := -1
	for i, co := range c.objects {
		if co.path == "" {
			continue
		}
		if co.lastAccessed.Before(lru) {
			lru = co.lastAccessed
			lruI = i
		}
	}
	if lruI == -1 {
		return cacheObject{}
	}
	co := c.objects[lruI]
	c.objects[lruI].path = ""
	c.numberOfObjects = 99
	c.size -= co.size
	return co
}

type cacheStat struct {
	count       int
	size        Size
	leastRU     time.Time // least recently used
	leastRUPath string
	mostRU      time.Time // most recently used
	mostRUPath  string
}

func (cs cacheStat) String() string {
	return fmt.Sprintf(`cacheStat
	count       %d
	size        %s
	leastRU     %s
	leastRUPath %s
	mostRU      %s
	mostRUPath  %s
`,
		cs.count,
		cs.size,
		cs.leastRU,
		cs.leastRUPath,
		cs.mostRU,
		cs.mostRUPath,
	)
}

func (c *cache) stat() cacheStat {
	size := Size(0)
	count := 0
	mru := time.Time{}
	mruP := ""
	lru := time.Now().AddDate(100, 0, 0)
	lruP := ""

	for _, co := range c.objects {
		if co.path == "" {
			continue
		}
		size += co.size
		count++
		if co.lastAccessed.After(mru) {
			mru = co.lastAccessed
			mruP = co.path
		} else if co.lastAccessed.Before(lru) {
			lru = co.lastAccessed
			lruP = co.path
		}
	}

	cs := cacheStat{
		count:       count,
		size:        size,
		mostRU:      mru,
		mostRUPath:  mruP,
		leastRU:     lru,
		leastRUPath: lruP,
	}
	return cs
}

func (c *cache) clear() {
	for i := range c.objects {
		c.objects[i].path = ""
	}
}

func (c *cache) cacheByRules(r CacheRules) []string {
	paths := []string{}

	// p := c.clearBySize(r)
	// paths = append(paths, p...)

	p := c.clearByTime(r)
	paths = append(paths, p...)

	return paths
}

func (c *cache) clearBySize(r CacheRules) []string {
	paths := []string{}
	for c.size > r.MaxTotalCacheSize {
		co := c.delLRU()
		paths = append(paths, co.path)
	}
	return paths
}

func (c *cache) clearByTime(r CacheRules) []string {
	paths := []string{}
	for _, co := range c.objects {
		if co.lastAccessed.Before(time.Now().Add(-r.MaxTimeSinceUse)) {
			paths = append(paths, co.path)
			co.path = ""
		}
	}
	return paths
}

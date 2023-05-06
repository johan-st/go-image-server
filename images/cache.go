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

type CacheRules struct {
	MaxTimeSinceUse time.Duration // Max age in seconds		(default: 0, unlimited)
	MaxSize         Size          // Max cache size in bytes	(default: 1 Gigabyte)
}

type cache struct {
	totalSize       Size
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
`, c.totalSize, c.numberOfObjects, c.cap, len(c.objects))
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
		totalSize:       0,
		numberOfObjects: 0,
		cap:             capacity,
		objects:         make([]cacheObject, 0, capacity),
	}
}

func (c *cache) add(co cacheObject) {
	c.objects = append(c.objects, co)
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

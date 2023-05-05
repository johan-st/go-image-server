package images

import (
	"fmt"
	"time"
)

// ImageHandler will try to keep the cache within these limits but does not guarantee it.
//
// note: Use 0 (zero) to explicitly set default.

type CacheRules struct {
	MaxTimeSinceUse time.Duration // Max age in seconds		(default: 0, unlimited)
	MaxSize         Size          // Max cache size in bytes	(default: 1 Gigabyte)
}

type cache []cacheObject

type cacheObject struct {
	path         string
	size         Size
	lastAccessed time.Time
}

func (c *cache) add(co cacheObject) {
	*c = append(*c, co)
}

func (c *cache) get(path string) (cacheObject, error) {
	for i, o := range *c {
		if o.path == "" {
			continue
		}
		if o.path == path {
			(*c)[i].lastAccessed = time.Time{}
			return o, nil
		}
	}
	return cacheObject{}, fmt.Errorf("cache could not locate object with path %s", path)
}

func (c *cache) del(path string) {
	for i, co := range *c {
		if co.path == path {
			(*c)[i].path = ""
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

func (cs *cacheStat) String() string {
	return fmt.Sprintf("----- cache -----\n  - count: %d\n  - size: %d\n  - least recently used:\n     - path:%s\n     - time: %s\n  - most recently used:\n     - path:%s\n     - time: %s\n\n-----------------  \n\n",
		cs.count,
		cs.size,
		cs.leastRUPath,
		cs.leastRU,
		cs.mostRUPath,
		cs.mostRU,
	)
}

func (c *cache) stat() cacheStat {
	size := Size(0)
	count := 0
	mru := time.Time{}
	mruP := ""
	lru := time.Time{}
	lruP := ""
	for _, co := range *c {
		if co.path == "" {
			continue
		}
		size += co.size
		count++
		if co.lastAccessed.After(mru) {
			mru = co.lastAccessed
			mruP = co.path
			fmt.Println("mru", mru)
			fmt.Println("mruP", mruP)
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
	// DEBUG: FIX THIS FIRST
	// TODO: why not updated?
	fmt.Println("stat:", cs)
	return cs
}

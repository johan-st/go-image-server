package images

import (
	"fmt"
	"sort"
	"testing"
	"time"
)

func Test_cache(t *testing.T) {
	// arrange
	// creating a new cache
	c := newCache(4)

	deleted := []cacheObject{}
	added := []cacheObject{
		{"test-1", 500, time.Now().AddDate(-1, 0, 0)},
		{"test-2", 500, time.Now().AddDate(0, -1, 0)},
		{"test-3", 500, time.Now().AddDate(0, 0, -1)},
		{"test-0", 500, time.Now()},
	}

	// act
	for i, co := range added {
		c.add(co)
		fmt.Printf("add %d\n", i+1)

	}

	if !assert(added, deleted, c) {
		fmt.Println("fail assert", c)
		t.Fatal("state was inconsistent")
	}

	toDel := &added[len(added)-1]
	fmt.Println("pre del of "+toDel.path, c)
	c.del(toDel.path)
	deleted = append(deleted, *toDel)
	toDel.path = ""
	fmt.Println("post del of "+toDel.path, c)

	if !assert(added, deleted, c) {
		fmt.Println("fail assert 2", c)
		t.Fatal("state was inconsistent")
	}

}

// -----------------
// ---- HELPERS ----
// -----------------

// assert that state is consistant with expected state
func assert(added, removed []cacheObject, c cache) bool {
	// are added in cache?
	for i, co := range added {
		// check deleted
		if co.path == "" {
			continue
		}
		coGet, err := c.get(co.path)
		if err != nil {
			fmt.Printf("was checking added[%d] when... %s\n", i, err)
			return false
		}
		if !matchingCacheObjects(co, coGet) {
			fmt.Printf("was matching added[%d] against obGet when... %s\n", i, err)
			return false
		}
	}

	// try to find removed objects
	for i, co := range removed {
		// check deleted
		if co.path == "" {
			continue
		}
		coGet, err := c.get(co.path)
		if err == nil {
			fmt.Printf("was looking for removed[%d] in cache and found it.. \npath: %s\npath: %s\n\n", i, co.path, coGet.path)
			return false
		}
	}

	return true
}

func matchingCacheObjects(o1, o2 cacheObject) bool {
	if o1.size != o2.size {
		return false
	}
	if o1.path != o2.path {
		return false
	}
	// Get updates time here.
	// if o1.lastAccessed != o2.lastAccessed {
	// 	return false
	// }

	return true

}

func Test_cache_cacheByRules(t *testing.T) {

	// files := []string{"test-1", "test-2", "test-3", "test-0"}
	// arrange
	cache := newCache(4)
	cache.add(cacheObject{"test-1", 100, time.Now().AddDate(-1, 0, 0)})
	cache.add(cacheObject{"test-2", 200, time.Now().AddDate(0, -1, 0)})
	cache.add(cacheObject{"test-3", 400, time.Now().AddDate(0, 0, -1)})
	cache.add(cacheObject{"test-0", 800, time.Now()})

	// act
	crNoEvict := CacheRules{
		MaxTotalCacheSize: 0,
		MaxTimeSinceUse:   366 * 24 * time.Hour,
	}
	crEvictSize := CacheRules{
		MaxTotalCacheSize: 400,
		MaxTimeSinceUse:   0,
	}
	crEvictTime := CacheRules{
		MaxTotalCacheSize: 0,
		MaxTimeSinceUse:   25 * time.Hour,
	}
	// assert
	paths := cache.cacheByRules(crNoEvict)
	if len(paths) != 0 {
		t.Fatal("no eviction should have occured")
	}

	paths = cache.cacheByRules(crEvictSize)
	if len(paths) != 2 {
		for _, p := range paths {
			fmt.Println("evicted", p)
			fmt.Println(cache.stat())
		}
		t.Fatal("2 evictions should have occured")
	}
	for _, p := range paths {
		fmt.Println(p)
	}

	paths = cache.cacheByRules(crEvictTime)
	if len(paths) != 2 {
		t.Fatal("2 evictions should have occured")
	}
	for _, p := range paths {
		if p != "test-1" && p != "test-2" {
			t.Fatal("wrong files were evicted")
		}
	}
	for _, p := range paths {
		fmt.Println(p)
	}

}

func Test_cache_delLRU(t *testing.T) {
	// arrange
	cache := newCache(4)
	cos := []cacheObject{
		{"test-0", 800, time.Now()},
		{"test-1", 100, time.Now().AddDate(-1, 0, 0)},
		{"test-2", 200, time.Now().AddDate(0, -1, 0)},
		{"test-3", 400, time.Now().AddDate(0, 0, -1)},
	}

	for _, co := range cos {
		cache.add(co)
	}

	sort.Slice(cos, func(i, j int) bool {
		return cos[i].lastAccessed.Before(cos[j].lastAccessed)
	})

	// act
	// assert
	for _, coRef := range cos {
		co := cache.delLRU()
		if co.path != coRef.path {
			t.Fatalf("correct LRU not deleted, wanted %s, got %s", coRef.path, co.path)
		}
	}

}

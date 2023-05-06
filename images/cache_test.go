package images

import (
	"fmt"
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

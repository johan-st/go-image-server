package images

import (
	"fmt"
	"sync"
)

// TimeSource is an interface to facilitate test
// type TimeSource interface {
// 	Now() time.Time
// 	Since(time.Time) time.Duration
// }

// CacheObject represents a cached image.
// type CacheObject struct {
// 	Path            string
// 	ImageParameters ImageParameters
// 	Size            Size
// 	LastAccessed    time.Time
// }

// LRU is a bespoke Least Recently Used cache. It is used to keep track of
// image files does therefor not store any data. What is of interest is
// knowing wether a certain file is generated and the order in which the
// files are accessed. LRU is thread safe.
type lru struct {
	// timeSource    TimeSource
	cap           int
	len           int
	head          *node
	tail          *node
	lookup        map[string]*node
	reverseLookup map[*node]string
	trimChan      chan<- string
	lMutex        sync.RWMutex
	rlMutex       sync.RWMutex
}

// NewLru creates a new Lru cache with the given capacity and trimedPathsChan.
// The capacity is the maximum number of paths that can be stored in the
// cache.
//
// The trimedPathsChan is used to communicate which paths are trimmed from the cache. When a path is removed from the cache it will be sent to the trimedPathsChan. The caller is responsible for handling removal of cache-filesbased on the trimedPathsChan messages and aöso for closing the channel. The channel should have a buffer size of at least 1 but larger is higky recommended.The channel length should be monitored and if it is close to full a warning should be issued.
func newLru(cap int, trimedPathsChan chan<- string) *lru {
	return &lru{
		cap:           cap,
		lookup:        make(map[string]*node),
		reverseLookup: make(map[*node]string),
		trimChan:      trimedPathsChan,
	}
}

//	type lruWithStats struct {
//		lru      *Lru
// 		timeIntervals []time.Duration
//		accesses map[time.Duration]int // hits during given timaspans
//		misses   map[time.Duration]int // misses during given timaspans
//		mruTime  time.Time
//		lruTime  time.Time
//	}

type node struct {
	prev, next *node
	id         int
	path       string
}

func (l *lru) Contains(filepath string) bool {
	defer printLookups(l, fmt.Sprintf("contains(%s)", filepath))
	if _, ok := l.lookupNode(filepath); ok {
		return true
	} else {
		return false
	}
}

func (l *lru) AddOrUpdate(id int, filepath string) bool {
	defer printLookups(l, fmt.Sprintf("addOrUpdate(%d, %s)", id, filepath))

	// fmt.Println("DEBUG: lru.AddOrUpdate(id) filepath:", filepath, "id:", id)
	if n, ok := l.lookupNode(filepath); ok {
		// fmt.Println("DEBUG: lru.AddOrUpdate(id) moveToFront, n:", n)
		l.moveToFront(n)
		return true
	} else {
		// create new node
		n := &node{id: id, path: filepath}
		// fmt.Println("DEBUG: lru.AddOrUpdate(id) addToFront, n:", n)
		// set lookups
		l.addToLookup(n, filepath)
		// add to front
		l.addToFront(n)
		// trim if needed
		l.trim()

		return false
	}
}

func (l *lru) Delete(id int) int {
	defer printLookups(l, fmt.Sprintf("delete(%d)", id))
	fmt.Println("DEBUG: lru.Delete(id) id:", id)
	var (
		numDeleted int
		curr       *node
		next       *node
	)
	curr = l.head
	if curr != nil {
		next = curr.next
	}

	// fmt.Println("DEBUG: lru.Delete(id) head:", l.head)
	// fmt.Println("DEBUG: lru.Delete(id) next:", next)

	// fmt.Println("DEBUG: lru.Delete(id) id:", id)
	for curr != nil {
		// fmt.Print("DEBUG: \n____________ loop ____________\n")
		// fmt.Println("DEBUG: curr.id == id", curr.id == id)
		if curr.id == id {
			path, ok := l.lookupPath(curr)
			if !ok {
				// fmt.Println("DEBUG: lru.Delete(id) id:", id)
				// fmt.Println("DEBUG: lru.Delete(id) node:", curr)
				panic("lru.Remove(id): node not found in lookup")
			}

			// fmt.Println("DEBUG: lru.Delete(id) path ok:", path)
			l.detatchNode(curr)
			l.removeFromLookup(curr, path)
			numDeleted++

			l.trimChan <- path
		}
		curr = next
		if curr != nil {
			next = curr.next
		}
	}

	return numDeleted
}

// func (l *lru) LoadDir(dirpath string) error {
// 	// load all files in dirpath into cache
// 	//    if cache is full, trim it
// 	panic("not implemented")
// }

func (l *lru) trim() {
	for l.len > l.cap {
		node := l.tail
		key, _ := l.lookupPath(node)

		l.detatchTail()
		l.removeFromLookup(node, key)

		l.trimChan <- key
	}
	//     TODO: consider handling deletion in this function
}

// List operations

func (l *lru) addToFront(n *node) {
	if l.len == 0 {
		l.head = n
		l.tail = n
		l.len++
		return
	}
	n.next = l.head
	l.head.prev = n
	l.head = n
	l.len++

}

func (l *lru) moveToFront(n *node) {
	// check if n is front
	if l.head == n {
		return
	}

	// check if n is  tail
	if l.tail == n {
		l.tail = n.prev
	}

	// detatch from neighbours
	if n.prev != nil {
		n.prev.next = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	}

	// set node links
	n.prev = nil
	n.next = l.head

	// set current head to point new node
	l.head.prev = n

	// set new head
	l.head = n
}

func (l *lru) detatchTail() {
	n := l.tail

	// link tail to previous node
	l.tail = n.prev

	// tail next is nil
	l.tail.next = nil

	// detatch node from list
	n.prev = nil

	// decrement len
	l.len--
}

func (l *lru) detatchNode(n *node) {
	// link previous node to next node
	if n.prev != nil {
		n.prev.next = n.next
	}

	// link next node to previous node
	if n.next != nil {
		n.next.prev = n.prev
	}

	// remove node links
	n.prev = nil
	n.next = nil

	// decrement len
	l.len--

}

// Lookup operations

func (l *lru) lookupNode(path string) (*node, bool) {
	l.lMutex.RLock()
	n, ok := l.lookup[path]
	l.lMutex.RUnlock()
	return n, ok
}

func (l *lru) lookupPath(n *node) (string, bool) {
	l.rlMutex.RLock()
	path, ok := l.reverseLookup[n]
	l.rlMutex.RUnlock()
	return path, ok
}

func (l *lru) addToLookup(n *node, path string) {
	l.lMutex.Lock()
	l.rlMutex.Lock()
	l.lookup[path] = n
	l.reverseLookup[n] = path
	l.lMutex.Unlock()
	l.rlMutex.Unlock()

}

func (l *lru) removeFromLookup(n *node, path string) {
	l.lMutex.Lock()
	l.rlMutex.Lock()
	delete(l.lookup, path)
	delete(l.reverseLookup, n)
	l.lMutex.Unlock()
	l.rlMutex.Unlock()
}

func printLookups(l *lru, name string) {
	fmt.Println("DEBUG: ", name, "lookups:")
	for _, n := range l.lookup {
		fmt.Println("DEBUG: lookup:", n)
	}
	for _, p := range l.reverseLookup {
		fmt.Println("DEBUG: reverseLookup:", p)
	}
}

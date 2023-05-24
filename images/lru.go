package images

import (
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
}

func (l *lru) Access(filepath string) bool {
	if n, ok := l.lookupNode(filepath); ok {
		l.moveToFront(n)
		return true
	} else {
		// create new node
		n := &node{}
		// set lookups
		l.addToLookup(n, filepath)
		// add to front
		l.addToFront(n)
		// trim if needed
		l.trim()
		return false
	}
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

	// detatch node from list
	n.prev = nil

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

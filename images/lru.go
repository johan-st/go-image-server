package images

import "sync"

// type TimeSource interface {
// 	Now() time.Time
// 	Since(time.Time) time.Duration
// }

// Lru is a least recently used cache
// it is thread safe
// it is a stripped down version of a regular LRU since we only need to keep track of filepaths and the order they were accessed in
type Lru struct {
	// timeSource    TimeSource
	cap           int
	Len           int
	head          *node
	tail          *node
	lookup        map[string]*node
	reverseLookup map[*node]string
	trimChan      chan<- string
	lMutex        sync.RWMutex
	rlMutex       sync.RWMutex
}

func NewLru(cap int, trimedPathsChan chan<- string) *Lru {
	return &Lru{
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

func (l *Lru) Access(filepath string) bool {
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

// func (l *Lru) LoadDir(dirpath string) error {
// 	// load all files in dirpath into cache
// 	//    if cache is full, trim it
// 	panic("not implemented")
// }

func (l *Lru) trim() {
	for l.Len > l.cap {
		node := l.tail
		key, _ := l.lookupPath(node)

		l.detatchTail()
		l.removeFromLookup(node, key)

		l.trimChan <- key
	}
	//     TODO: consider handling deletion in this function
}

// List operations

func (l *Lru) addToFront(n *node) {
	if l.Len == 0 {
		l.head = n
		l.tail = n
		l.Len++
		return
	}
	n.next = l.head
	l.head.prev = n
	l.head = n
	l.Len++

}

func (l *Lru) moveToFront(n *node) {
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

func (l *Lru) detatchTail() {
	n := l.tail

	// link tail to previous node
	l.tail = n.prev

	// detatch node from list
	n.prev = nil

	// decrement len
	l.Len--
}

// Lookup operations

func (l *Lru) lookupNode(path string) (*node, bool) {
	l.lMutex.RLock()
	n, ok := l.lookup[path]
	l.lMutex.RUnlock()
	return n, ok
}

func (l *Lru) lookupPath(n *node) (string, bool) {
	l.rlMutex.RLock()
	path, ok := l.reverseLookup[n]
	l.rlMutex.RUnlock()
	return path, ok
}

func (l *Lru) addToLookup(n *node, path string) {
	l.lMutex.Lock()
	l.rlMutex.Lock()
	l.lookup[path] = n
	l.reverseLookup[n] = path
	l.lMutex.Unlock()
	l.rlMutex.Unlock()
}

func (l *Lru) removeFromLookup(n *node, path string) {
	l.lMutex.Lock()
	l.rlMutex.Lock()
	delete(l.lookup, path)
	delete(l.reverseLookup, n)
	l.lMutex.Unlock()
	l.rlMutex.Unlock()
}

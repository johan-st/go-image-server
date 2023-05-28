package images

import (
	"testing"
)

func TestLru(t *testing.T) {
	t.Parallel()

	trimChan := make(chan string, 100)
	lru := newLru(3, trimChan)

	lt := lruT{
		t:        t,
		lru:      lru,
		trimChan: trimChan,
	}

	lt.miss("a") // a
	lt.hit("a")  // a
	lt.miss("b") // b a
	lt.miss("c") // c b a
	lt.miss("d") // d c b
	lt.miss("e") // e d c	-	trim a
	lt.miss("f") // f e d	-	trim a b
	lt.miss("g") // g f e	-	trim a b c
	lt.miss("h") // h g f	-	trim a b c d e
	lt.hit("h")  // h g f	-	trim a b c d e
	lt.hit("f")  // f h g	-	trim a b c d e
	lt.hit("g")  // g f h	-	trim a b c d e
	lt.miss("a") // a g f	-	trim a b c d e h

	lt.trimed("a")
	lt.trimed("b")
	lt.trimed("c")
	lt.trimed("d")
	lt.trimed("e")
	lt.trimed("h")

	lt.noTrim()
}

// HELPER
type lruT struct {
	lru      *lru
	t        *testing.T
	trimChan <-chan string
}

func (l *lruT) miss(s string) {
	l.t.Helper()
	if l.lru.AddOrUpdate(s) {
		l.t.Errorf("AddOrUpdate(\"%s\") expected miss", s)
	}
	l.t.Log("miss: ", s)
}

func (l *lruT) hit(s string) {
	l.t.Helper()
	if !l.lru.AddOrUpdate(s) {
		l.t.Errorf("AddOrUpdate(\"%s\") expected hit", s)
	}
	l.t.Log("hit:  ", s)
}

func (l *lruT) trimed(s string) {
	l.t.Helper()
	trim := <-l.trimChan
	if trim != s {
		l.t.Errorf("trim expected %s got %s", s, trim)
	}
	l.t.Log("trim: ", trim)

}

func (l *lruT) noTrim() {
	l.t.Helper()
	for {
		select {
		case v := <-l.trimChan:
			l.t.Fatalf("no trim expected but got %s", v)
		default:
			return
		}
	}
}

// func printChan(c <-chan string) {
// 	for {
// 		select {
// 		case v := <-c:
// 			println("channel got: ", v)
// 		default:
// 			return
// 		}
// 	}
// }

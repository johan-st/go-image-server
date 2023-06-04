package images

import (
	"testing"
)

func TestLruAdd(t *testing.T) {
	t.Parallel()

	trimChan := make(chan string, 100)
	lru := newLru(3, trimChan)

	lt := lruT{
		t:        t,
		lru:      lru,
		trimChan: trimChan,
	}

	lt.miss(1, "a") // a1
	lt.miss(1, "b") // b1 a1
	lt.miss(1, "c") // c1 b1 a1

	lt.hit(1, "a") // a1 c1 b1
	lt.hit(1, "b") // b1 a1 c1
	lt.hit(1, "c") // c1 b1 a1

}

func TestLruTrim(t *testing.T) {
	t.Parallel()

	trimChan := make(chan string, 100)
	lru := newLru(3, trimChan)

	lt := lruT{
		t:        t,
		lru:      lru,
		trimChan: trimChan,
	}

	lt.miss(1, "a") // a1
	lt.miss(1, "b") // b1 a1
	lt.miss(1, "c") // c1 b1 a1
	lt.miss(1, "d") // d1 c1 b1
	lt.trimed("a")
	lt.noTrim()

}

func TestLruRemove(t *testing.T) {
	t.Parallel()

	trimChan := make(chan string, 100)
	lru := newLru(3, trimChan)

	lt := lruT{
		t:        t,
		lru:      lru,
		trimChan: trimChan,
	}

	lt.miss(1, "a") // a1
	lt.miss(1, "b") // b1 a1
	lt.miss(1, "c") // c1 b1 a1

	lt.rm(1)
	lt.trimed("c")
	lt.trimed("b")
	lt.trimed("a")

	lt.miss(1, "a") // a1
	lt.miss(2, "b") // b1 a1
	lt.miss(2, "c") // c1 b1 a1

	lt.rm(2)
	lt.trimed("c")
	lt.trimed("b")

	lt.hit(1, "a") // a1

}

func TestLru(t *testing.T) {
	t.Parallel()

	trimChan := make(chan string, 100)
	lru := newLru(3, trimChan)

	lt := lruT{
		t:        t,
		lru:      lru,
		trimChan: trimChan,
	}

	lt.miss(1, "a") // a1
	lt.hit(1, "a")  // a1
	lt.miss(1, "b") // b1 a1
	lt.miss(1, "c") // c1 b1 a1
	lt.miss(1, "d") // d1 c1 b1

	lt.trimed("a")

	lt.miss(2, "e") // e2 d1 c1
	lt.trimed("b")

	lt.miss(2, "f") // f2 e2 d1
	lt.trimed("c")

	lt.miss(2, "g") // g2 f2 e2
	lt.trimed("d")

	lt.miss(2, "h") // h2 g2 f2
	lt.trimed("e")

	lt.hit(2, "h") // h2 g2 f2
	lt.hit(2, "f") // f2 h2 g2
	lt.hit(2, "g") // g2 f2 h2

	lt.miss(1, "a") // a1 g2 f2
	lt.trimed("h")

	lt.noTrim()

	lt.rm(2) // a1
	lt.trimed("g")
	lt.trimed("f")
	lt.noTrim()

	lt.miss(2, "g") // g2 a1
	lt.noTrim()

	lt.miss(2, "f") // f2 g2 a1
	lt.noTrim()

	lt.rm(1) // f2 g2
	lt.trimed("a")
	lt.noTrim()

}

// HELPER
type lruT struct {
	lru      *lru
	t        *testing.T
	trimChan <-chan string
}

func (l *lruT) miss(id int, s string) {
	l.t.Helper()
	if l.lru.AddOrUpdate(id, s) {
		l.t.Errorf("Error: AddOrUpdate(\"%s\") expected miss", s)
	}
	l.t.Log("miss: ", s)
}

func (l *lruT) hit(id int, s string) {
	l.t.Helper()
	if !l.lru.AddOrUpdate(id, s) {
		l.t.Errorf("Error: AddOrUpdate(\"%s\") expected hit", s)
	}
	l.t.Log("hit:  ", s)
}

func (l *lruT) rm(id int) {
	// l.t.Helper()
	len := l.lru.len
	num := l.lru.Delete(id)
	if num == 0 {
		l.t.Errorf("Error: Delete(%d) expected to delete something", id)
	}
	l.t.Log("del:  ", id, "num: ", num, "\tlen: ", len, "->", l.lru.len)
}

func (l *lruT) trimed(s string) {
	l.t.Helper()
	trim := <-l.trimChan
	if trim != s {
		l.t.Errorf("Error: trim expected %s got %s", s, trim)
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
			l.t.Log("no trim")
			return
		}
	}
}

package pkg1

import (
	"sync"
)

// Type1 is a test struct.
type Type1 struct {
	m sync.RWMutex
	// synt: m.Lock
	i int
	// synt: m.RLock
	j int64

	mut sync.Mutex
	// synt: mut.Lock
	k, l float64
}

// type block comment
type (
	// Type2
	Type2 struct {
		a int
	}
	// Type3
	Type3 struct {
		Type2
	}
	EmptyType struct {
	}
)

// synt: !t.m.Lock
func (t *Type1) func1() {
	t.m.Lock()
	t.func2()
}

// synt: t.m.Lock
func (t Type1) func2() {

}

func freeFunc() {

}

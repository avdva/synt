package pkg1

import (
	"sync"
)

// Type1 is a test struct.
type Type1 struct {
	m sync.RWMutex
	// synt: m:L
	i int
	// synt: m:R
	j int64

	mut sync.Mutex
	// synt: mut:L
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

// synt: t.m:L
func (t *Type1) func1() {

}

// synt: t.m:L
func (t Type1) func2() {

}

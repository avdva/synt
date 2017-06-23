package pkg1

import (
	"sync"
)

// Type1 is a test struct.
// synt: i:m:L, j:m:R, k:mut:L
type Type1 struct {
	m   sync.RWMutex
	mut sync.Mutex

	i int
	j int64
	k float64
}

// type block comment
type (
	// Type2
	Type2 struct {
		a int
	}
	// Type3
	Type3 struct {
		a int
	}
)

// synt: t.m:L
func (t *Type1) func1() {

}

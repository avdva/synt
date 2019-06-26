package optest

import (
	"sync"
)

func getInt() int {
	return 0
}

func getInt2(a1 int) int {
	return 0
}

func getPtr(arg int) *int {
	return &arg
}

type str struct {
	m sync.RWMutex
}

func f1() {
	var i int
	var sl []int
	var s str
	a := 5
	sl[i] = 8
	sl[i] = a
	b, c := a, sl[i]
	a = sl[getInt()]
	a, a = b, c
	a = getInt2(2)
	a = *getPtr(a)
	*getPtr(3) = 8
	s.m.Lock()
}

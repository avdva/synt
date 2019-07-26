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

func getSlice(_ struct{}) []int {
	return nil
}

func getSlice2(_ struct{}) []int {
	return nil
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
	sl[sl[i]] = getSlice2(struct{}{})[getSlice(struct{}{})[0]]
	*getPtr(3) = getInt2(a)
	a++
	b, c := a, sl[i]
	a = b + (c * b)
	a = sl[getInt()]
	a = getInt2(2)
	a = *getPtr(c)
	getSlice(struct{}{})[*getPtr(3)] = b
	s.m.Lock()
}

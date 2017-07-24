package pkg1

import (
	"sync"
)

// synt:t.m.RLock, t.mut.Lock
func (t *Type1) func3() int {
	t.m.Lock()
	return 3
}

func (t *Type1) func3_1(arg int) {
	t.func3()
	return
}

// synt:t.m.Lock
func (t *Type1) func3_2() {
}

func (t *Type1) func3_3() {
	t.m.RLock()
	t.func3_2()
}

// synt:t.m.RLock
func (t *Type1) func3_4() {
	t.func3_2()
}

func (t *Type1) func3_5() {
	t.m.RUnlock()
	t.m.Unlock()
}

func (t *Type1) func3_6() {
	t.m.Lock()
	t.m.Unlock()
	t.m.Unlock()
}

func (t *Type1) func4(arg int) int {
	t.func2()
	return 4
}

func (t *Type1) getM() *sync.RWMutex {
	return &t.m
}

func (t *Type1) func5() {
	a := 0
	{
		a = 1
		_ = a
		go func() {
			t.m.Lock()
		}()
	}
	t.m.Lock()
	switch {
	case a == 0:
	default:
		t.m.Unlock()
	}
	t.func1()
	if a == 5 {
		if a > 0 {
			t.m.RLock()
		}
	} else if a == 6 {
		defer t.m.RUnlock()
	} else {
		func() {
			t.m.Lock()
		}()
	}
	freeFunc()
	var b int
	a, b = 7, t.func4(t.func3())
	c := b == 0 && a == 1
	a += t.func3()
	select {
	case <-chan byte(nil):
		b++
	default:
		_ = c
		t.m.RUnlock()
	}
	t.getM().RLock()
	t.i++
}

func (t *Type1) func6() {
	go func() {
		t.func3()
	}()
	a := func() int {
		return 0
	}()
	_ = a
}

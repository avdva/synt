package pkg1

import (
	"sync"
)

// synt:t.m.RLock, t.mut.Lock
func (t *Type1) func3() int {
	t.m.Lock()
	return 3
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

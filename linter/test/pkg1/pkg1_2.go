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
	var v int
	var (
		v1 int
	)
	_, _ = v, v1
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
	freeFunc()
	t.func1()
	go t.func3_1(6)
	go t.func3_2()
	go func() {
		t.func3()
	}()
	t.self(0).ptr.self(t.func3()).getM().RLock()
}

func (t *Type1) func7() {
	a := func(val float64) int {
		t.m.Lock()
		t.func3()
		return 0
	}(t.k)
	_ = a
}

func (t *Type1) func8() {
	a := 0
	if a == 0 {
		t.m.RLock()
	} else if a == 1 {
		a = 5
	} else {
		t.m.Lock()
		t.m.Lock()
	}
	t.func3_4()
	t.m.Unlock()
}

func (t *Type1) func9() {
	a, b := 0, 0
	t.m.Lock()
	t.m.Unlock()
	if a == 0 {
		if b == 3 {
			t.m.Lock()
		} else {
			t.m.Lock()
		}
	} else if a == 4 {
		t.m.Lock()
	} else {
		t.m.Lock()
	}
	if true {
		t.mut.Lock()
	} else {
		t.mut.Unlock()
	}
	t.func3()
}

func (t *Type1) func10() {
	a := 0
	if a == 0 {
	} else {
		t.m.Lock()
	}
	t.func3_2()
}

func (t *Type1) func11() {
	a := 0
	if a == 0 {
		t.m.Lock()
		return
	} else if a == 1 {
		t.m.Lock()
		panic("dsf")
	} else {
		t.m.Lock()
	}
	t.func3_2()
}

func (t *Type1) func12() {
	t.m.Lock()
	defer t.m.Unlock()
	defer t.m.Unlock()
	defer func() {
		t.m.Lock()
	}()
}

func (t *Type1) func13() {
	t.m.Lock()
	func() {
		defer t.m.Unlock()
	}()
	defer t.m.Lock()
}

func (t *Type1) func14() {
	a := 0
	t.m.Lock()
	defer t.m.Unlock()
	if a == 0 {
		defer t.m.Unlock()
	} else if a == 1 {
		defer t.m.RUnlock()
	} else {
		defer func() {
			defer t.m.RUnlock()
			t.m.RUnlock()
		}()
	}
}

type f3 struct{}

func (f f3) func3() {

}

func (f f3) func3_2() {

}

func (t *Type1) func15() {
	{
		var t f3
		t.func3()
	}
	for t := range (chan f3)(nil) {
		t.func3_2()
	}
	if t := new(f3); t != nil {
		t.func3_2()
	}
	{
		t := f3{}
		t.func3()
	}
}

package pkg1

// synt:@m:L
func (t *Type1) func3() {
	t.m.Lock()
}

func (t *Type1) func4() {
	t.func2()
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

	}
	select {
	case <-(chan byte)(nil):
	default:
		t.m.RUnlock()
	}
}

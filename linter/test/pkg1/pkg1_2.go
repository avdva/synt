package pkg1

// synt:@m:L
func (t *Type1) func3() {
	t.m.Lock()
}

func (t *Type1) func4() {
	t.func2()
}

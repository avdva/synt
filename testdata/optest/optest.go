package optest

func getInt() int {
	return 0
}

func getInt2(a1 int) int {
	return 0
}

func getPtr(arg int) *int {
	return &arg
}

func f1() {
	var i int
	var sl []int
	a := 5
	sl[i] = 8
	sl[i] = a
	b, c := a, sl[i]
	a = sl[getInt()]
	a, a = b, c
	a = getInt2(2)
	a = *getPtr(a)
	*getPtr(3) = 8
}

package optest

func getInt() int {
	return 0
}

func f1() {
	var i int
	var sl []int
	a := 5
	sl[i] = 8
	b, c := a, sl[i]
	a = sl[getInt()]
	a, a = b, c
}

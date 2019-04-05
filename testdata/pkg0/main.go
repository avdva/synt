package main

import (
	"fmt"
	"sync"
)

var (
	a int
	m sync.Mutex
	// synt: m.Lock
	b int
)

func init() {
	b = 0
	b++
	fmt.Println(b)
}

func someFunc() {
	loc := 0
	_ = loc
	m.Lock()
	b = 0
	fmt.Println(b)
	m.Unlock()
	c = 0
	localFunc := func() {

	}
	localFunc()
}

func main() {
	var b int
	_ = b
	someFunc()
}

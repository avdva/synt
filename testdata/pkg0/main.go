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

func doubleLock() {
	m.Lock()
	m.Lock()
}

func doubleUnlock() {
	m.Lock()
	m.Unlock()
	m.Unlock()
}

func unlockedUnlock() {
	m.Unlock()
}

func main() {
	var b int
	_ = b
}

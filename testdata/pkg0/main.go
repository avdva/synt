package main

import (
	"fmt"
	"sync"
)

var (
	a int
	m sync.RWMutex
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

func ifLock() {
	m.Lock()
	if a := 0; a > 0 {
		m.RLock()
	}
}

func ifLock2() {
	if a := 0; a > 0 {
		m.Lock()
	} else {
		m.Lock()
	}
	m.Unlock()
}

func ifLock3() {
	if a := 0; a > 0 {
		m.Lock()
	} else {
		m.RLock()
	}
	m.Unlock()
}

func ifLock4() {
	if true {
		m.Lock()
	} else if false {
		m.Lock()
		if true {
			m.Unlock()
		} else {
			m.Unlock()
		}
		m.Lock()
	} else {
		m.RLock()
		m.RUnlock()
		m.Lock()
	}
	m.Unlock()
}

func main() {
	var b int
	_ = b
}

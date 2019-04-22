package main

import (
	"fmt"
	"sync"
)

var (
	a int
	m sync.RWMutex
	// synt: m.Lock wm.wmMut.Lock
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

func defferedUnlock() {
	defer m.Unlock()
}

func defferedDoubleUnlock() {
	m.Lock()
	defer m.Unlock()
	a = 0
	defer m.Unlock()
}

func defferedIfUnlock() {
	m.Lock()
	defer m.Unlock()
	defer m.Lock()
	if true {
		defer m.RUnlock()
	} else if false {
		defer m.Unlock()
	} else {
		defer m.Unlock()
	}
}

func guardedAccess() {
	loc := c
	_ = loc
	c = 5
	m.Lock()
	c = 7
	m.Unlock()
	m.RLock()
	loc = c
	m.RUnlock()
}

func main() {
	var b int
	_ = b
}

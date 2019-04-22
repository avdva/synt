package main

import (
	"sync"
)

var (
	// synt: m.Lock
	c int

	wm withMutex

	//synt: wm.wmMut.Lock
	n = 0

	em embeddedMutex

	//synt:em.Lock
	e float64
)

type withMutex struct {
	wmMut sync.RWMutex
}

type embeddedMutex struct {
	sync.RWMutex
}

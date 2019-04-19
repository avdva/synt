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
)

type withMutex struct {
	wmMut sync.RWMutex
}

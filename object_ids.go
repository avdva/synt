// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"strconv"
)

type ids struct {
	data map[interface{}]int
	last int
}

func newIds() *ids {
	return &ids{
		data: make(map[interface{}]int),
	}
}

func (i *ids) add(object interface{}) int {
	id, found := i.data[object]
	if found {
		return id
	}
	i.last++
	i.data[object] = i.last
	return i.last
}

func (i *ids) id(object interface{}) int {
	return i.data[object]
}

func (i *ids) strID(object interface{}) string {
	id, found := i.data[object]
	if !found {
		return ""
	}
	return strconv.Itoa(id)
}

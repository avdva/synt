// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"reflect"
	"strings"
)

type opchain []operation

func (oc opchain) String() string {
	var parts []string
	for _, o := range oc {
		if len(o.String()) == 0 {
			panic(reflect.TypeOf(o).String())
		}
		parts = append(parts, o.String())
	}
	return strings.Join(parts, "+")
}

func (oc opchain) ObjectName() string {
	if len(oc) == 1 {
		if wop, ok := oc[0].(*wOp); ok {
			return wop.ObjectName()
		}
	}
	var parts []string
	for _, o := range oc {
		parts = append(parts, o.ObjectName())
	}
	return strings.Join(parts, ".")
}

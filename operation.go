// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"bytes"
	"go/ast"
)

const (
	opRead opType = iota
	opWrite
	opExec
)

type opType int

func (t opType) String() string {
	switch t {
	case opRead:
		return "read"
	case opWrite:
		return "write"
	case opExec:
		return "exec"
	default:
		return "unknown"
	}
}

type op struct {
	typ    opType
	object *ast.Ident
}

func (o op) GoString() string {
	return o.typ.String() + ":" + o.object.Name
}

type opchain []op

func (oc opchain) GoString() string {
	var buff bytes.Buffer
	for i, o := range oc {
		if i != 0 {
			buff.WriteString(",")
		}
		buff.WriteString(o.GoString())
	}
	return buff.String()
}

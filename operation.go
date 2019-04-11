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
	buf := bytes.NewBufferString(o.typ.String() + " ")
	buf.WriteString(o.object.Name)
	return buf.String()
}

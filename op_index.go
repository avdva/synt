// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import "go/token"

type indexOp struct {
	x     opchain
	index opchain
}

func (op indexOp) Pos() token.Pos {
	return op.x[len(op.x)-1].Pos()
}

func (op indexOp) Type() opType {
	return opExec
}

func (op indexOp) ObjectName() string {
	return op.x.ObjectName() + "[" + op.index.ObjectName() + "]"
}

func (op indexOp) String() string {
	return op.Type().String() + ":" + op.ObjectName()
}

func (op indexOp) flatten() []opchain {
	return []opchain{op.index, op.x}
}

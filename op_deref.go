// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/token"
)

type derefOp struct {
	x opchain
}

func (op derefOp) Type() opType {
	return opExec
}

func (op derefOp) Pos() token.Pos {
	return op.x[len(op.x)-1].Pos()
}

func (op derefOp) ObjectName() string {
	return "*(" + op.x.ObjectName() + ")"
}

func (op derefOp) String() string {
	return op.Type().String() + ":" + "*(" + op.x.String() + ")"
}

// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import "go/token"

type wOp struct {
	lhs opchain
	rhs opchain
}

func (op wOp) Pos() token.Pos {
	return token.NoPos
}

func (op *wOp) Type() opType {
	return opW
}

func (o wOp) ObjectName() string {
	return o.lhs.ObjectName()
}

func (o wOp) String() string {
	result := o.Type().String() + ":"
	result += "(" + o.lhs.ObjectName() + "=" + o.rhs.ObjectName() + ")"
	return result
}

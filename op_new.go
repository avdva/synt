// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
)

type newOp struct {
	typ   ast.Expr
	inits []opchain
}

func (op newOp) Type() opType {
	return opExec
}

func (o newOp) ObjectName() string {
	return "new(" + exprName(o.typ) + ")"
}

func (o newOp) String() string {
	return o.Type().String() + ":" + o.ObjectName()
}

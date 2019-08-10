// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import "go/ast"

type rOp struct {
	expr ast.Expr
}

func newROpIdent(ident *ast.Ident) *rOp {
	return &rOp{expr: ident}
}

func newROpBasicLit(lit *ast.BasicLit) *rOp {
	return &rOp{expr: lit}
}

func (op *rOp) Type() opType {
	return opR
}

func (o rOp) ObjectName() string {
	return exprName(o.expr)
}

func (o rOp) String() string {
	result := o.Type().String() + ":" + o.ObjectName()
	return result
}

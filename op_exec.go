// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
	"go/token"
	"strings"
)

type execOp struct {
	fun  ast.Expr
	args []opchain
}

func (op *execOp) Type() opType {
	return opExec
}

func (o execOp) ObjectName() string {
	return exprName(o.fun)
}

func (op execOp) Pos() token.Pos {
	return op.fun.Pos()
}

func (o execOp) String() string {
	result := o.Type().String() + ":" + o.ObjectName()
	if len(o.args) > 0 {
		var args []string
		for _, arg := range o.args {
			args = append(args, arg.ObjectName())
		}
		result += "(" + strings.Join(args, ",") + ")"
	}
	return result
}

func (op execOp) flatten() []opchain {
	return op.args
}

type flattenable interface {
	flatten() []opchain
}

func flatten(chain opchain, level int) opchain {
	var result opchain
	for _, op := range chain {
		if level != 0 {
			newLevel := level
			if newLevel > 0 {
				newLevel--
			}
			if fl, ok := op.(flattenable); ok {
				for _, ch := range fl.flatten() {
					result = append(result, flatten(ch, newLevel)...)
				}
			}
		}
		result = append(result, op)
	}
	return result
}

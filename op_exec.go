// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
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
	result := exprName(o.fun)
	if len(o.args) > 0 {
		var args []string
		for _, arg := range o.args {
			args = append(args, arg.ObjectName())
		}
		result += "(" + strings.Join(args, ",") + ")"
	}
	return result
}

func (o execOp) String() string {
	return o.Type().String() + ":" + o.ObjectName()
}

func flatten(chain opchain, level int) opchain {
	var result opchain
	for _, op := range chain {
		if level != 0 {
			newLevel := level
			if newLevel > 0 {
				newLevel--
			}
			switch typed := op.(type) {
			case *execOp:
				for _, arg := range typed.args {
					result = append(result, flatten(arg, newLevel)...)
				}
			}
		}
		result = append(result, op)
	}
	return result
}

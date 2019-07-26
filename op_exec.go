// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
	"strconv"
)

var (
	accessFuncs map[string]struct {
		formatter func(execOp) string
	}
)

func init() {
	accessFuncs = map[string]struct {
		formatter func(execOp) string
	}{
		"__indexaccess": {
			formatter: func(op execOp) string {
				if len(op.args) != 2 {
					panic("bad access func")
				}
				return op.args[0].ObjectName() + "[" + op.args[1].ObjectName() + "]"
			},
		},
		"__ptraccess": {
			formatter: func(op execOp) string {
				if len(op.args) != 1 {
					panic("bad access func")
				}
				return "(*" + op.args[0].ObjectName() + ")"
			},
		},
	}
}

type execOp struct {
	fun  ast.Expr
	args []opchain
}

func (op *execOp) typ() opType {
	return opExec
}

func (o execOp) ObjectName() string {
	var result string
	name := exprName(o.fun)
	if info, found := accessFuncs[name]; found {
		result = info.formatter(o)
	} else {
		result += name
		if len(o.args) > 0 {
			result += "(" + strconv.Itoa(len(o.args)) + ")"
		}
	}
	return result
}

func (o execOp) String() string {
	return o.typ().String() + ":" + o.ObjectName()
}

func isAccessFunc(name string) bool {
	_, found := accessFuncs[name]
	return found
}

func cleanAccessExpr(chain opchain) opchain {
	result := make(opchain, len(chain))
	copy(result, chain)
	var iter int
	for i := 0; i < len(result); {
		iter++
		l := result[i]
		if exec, ok := l.(*execOp); ok {
			println(iter, " ", i)
			_, found := accessFuncs[exprName(exec.fun)]
			if found {
				if i-len(exec.args) < 0 {
					panic(exprName(exec.fun))
				}
				result = append(result[:i-len(exec.args)], append([]operation{l}, result[i+1:]...)...)
				i -= (len(exec.args) - 1)
				continue
			}
		}
		i++
	}
	println("----")
	return result
}

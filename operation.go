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
	typ opType

	object *ast.Ident
	args   []ast.Expr
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

func statementsToOpchain(statements []ast.Stmt) []opchain {
	var result []opchain
	for _, statement := range statements {
		switch typed := statement.(type) {
		case *ast.AssignStmt:
			for _, rhs := range typed.Rhs {
				result = append(result, expandLhs(rhs))
			}
			for _, lhs := range typed.Lhs {
				result = append(result, expandExpr(lhs))
			}
		case *ast.ExprStmt:
			result = append(result, exprToOpChain(typed))
		}
	}
	return result
}

func expandLhs(lhs ast.Expr) opchain {
	chain := expandExpr(lhs)
	return chain
}

func exprToOpChain(expr *ast.ExprStmt) opchain {
	var result opchain
	switch typed := expr.X.(type) {
	case *ast.CallExpr:
		result = append(result, expandExpr(typed)...)
	}
	return result
}

func expandExpr(expr ast.Expr) opchain {
	var result opchain
	for expr != nil {
		switch typed := expr.(type) {
		case *ast.CallExpr:
			switch fTyped := typed.Fun.(type) {
			case *ast.Ident:
				for _, arg := range typed.Args {
					result = append(result, expandExpr(arg)...)
				}
				result = append([]op{{object: fTyped, args: typed.Args, typ: opExec}}, result...)
				expr = nil
			case *ast.SelectorExpr:
				for _, arg := range typed.Args {
					result = append(result, expandExpr(arg)...)
				}
				result = append([]op{{object: fTyped.Sel, args: typed.Args, typ: opExec}}, result...)
				expr = fTyped.X
			}
		case *ast.SelectorExpr:
			result = append([]op{{object: typed.Sel, typ: opRead}}, result...)
			expr = typed.X
		case *ast.Ident:
			result = append([]op{{object: typed, typ: opRead}}, result...)
			expr = nil
		default:
			expr = nil
		}
	}
	return result
}

// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"bytes"
	"fmt"
	"go/ast"
	"reflect"
	"strconv"
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

	object ast.Expr
	args   []ast.Expr
}

func (o op) objName() string {
	switch typed := o.object.(type) {
	case *ast.Ident:
		return typed.Name
	}
	return ""
}

func (o op) String() string {
	result := o.typ.String() + ":"
	if o.object != nil {
		name := o.objName()
		if len(name) == 0 {
			name = "<expr>"
		}
		result += name
	}
	if len(o.args) > 0 {
		switch o.typ {
		case opExec:
			result += "(" + strconv.Itoa(len(o.args)) + ")"
		}
	}
	return result
}

type opchain []op

func (oc opchain) String() string {
	var buff bytes.Buffer
	for i, o := range oc {
		if i != 0 {
			buff.WriteString("+")
		}
		buff.WriteString(o.String())
	}
	return buff.String()
}

type opFlow []opchain

func (of opFlow) String() string {
	var buff bytes.Buffer
	for i, o := range of {
		if i != 0 {
			buff.WriteString("->")
		}
		if o == nil {
			buff.WriteString("<nil>")
		} else {
			buff.WriteString(o.String())
		}
	}
	return buff.String()
}

func statementsToOpchain(statements []ast.Stmt) opFlow {
	var result opFlow
	for _, statement := range statements {
		switch typed := statement.(type) {
		case *ast.AssignStmt:
			result = append(result, expandAssignment(typed.Lhs, typed.Rhs)...)
		case *ast.ExprStmt:
			result = append(result, expandExprEtmt(typed))
		default:
			fmt.Printf("statementsToOpchain: skipping %s\n", reflect.ValueOf(statement).Type().String())
		}
	}
	return result
}

func expandAssignment(lhs, rhs []ast.Expr) opFlow {
	var result opFlow
	for _, rhs := range rhs {
		result = append(result, expandExpr(rhs))
	}
	for _, lhs := range lhs {
		expanded := expandExpr(lhs)
		result = append(result, expanded)

	}
	return result
}

func expandExprEtmt(expr *ast.ExprStmt) opchain {
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
		case *ast.IndexExpr:
			result = append(result, expandExpr(typed.Index)...)
			result = append(result, expandExpr(typed.X)...)
			expr = nil
		case *ast.BasicLit:
			result = append(result, op{typ: opRead, args: []ast.Expr{typed}})
			expr = nil
		default:
			fmt.Printf("expandExpr: skipping %s\n", reflect.ValueOf(expr).Type().String())
			expr = nil
		}
	}
	return result
}

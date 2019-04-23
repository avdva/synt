// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"bytes"
	"fmt"
	"go/ast"
	"reflect"
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
	typ    opType
	object *ast.Ident
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
		fmt.Println(reflect.TypeOf(statement).String())
		switch typed := statement.(type) {
		case *ast.AssignStmt:

		case *ast.ExprStmt:
			result = append(result, exprToOpChain(typed))
		}
	}
	return result
}

func exprToOpChain(expr *ast.ExprStmt) opchain {
	var result opchain
	switch typed := expr.X.(type) {
	case *ast.CallExpr:
		// TODO(avd) - check args for non-deffered calls.
		result = append(result, callExprToOpChain(typed)...)
	}
	return result
}

func callExprToOpChain(expr ast.Expr) opchain {
	var result opchain
	for _, elem := range expandCallExpr(expr) {
		typ := opRead
		if elem.call {
			typ = opExec
		}
		result = append(result, op{typ: typ, object: elem.id})
	}
	return result
}

type callChain []callChainElem

type callChainElem struct {
	id   *ast.Ident
	call bool
	args []ast.Expr
}

func expandCallExpr(expr ast.Expr) callChain {
	var result callChain
	for expr != nil {
		switch typed := expr.(type) {
		case *ast.CallExpr:
			switch fTyped := typed.Fun.(type) {
			case *ast.Ident:
				result = append([]callChainElem{{id: fTyped, args: typed.Args, call: true}}, result...)
				expr = nil
			case *ast.SelectorExpr:
				result = append([]callChainElem{{id: fTyped.Sel, args: typed.Args, call: true}}, result...)
				expr = fTyped.X
			}
		case *ast.SelectorExpr:
			result = append([]callChainElem{{id: typed.Sel}}, result...)
			expr = typed.X
		case *ast.Ident:
			result = append([]callChainElem{{id: typed}}, result...)
			expr = nil
		default:
			expr = nil
		}
	}
	return result
}

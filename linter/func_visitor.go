// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"go/ast"
	"reflect"
)

type funcVisitor struct {
}

func (fv *funcVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	println(reflect.TypeOf(node).String())
	switch typed := node.(type) {
	case *ast.GoStmt:
		return nil
	case *ast.CallExpr:
		println("  ", reflect.TypeOf(typed.Fun).String())
		//*ast.CommClause
		return nil
	case *ast.IfStmt:
		return nil
	}
	return fv
}

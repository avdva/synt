// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"go/ast"
	"reflect"
)

const (
	mutStateUnlocked = iota
	mutStateL
	mutStateR
	mutStateMayL
	mutStateMayR
	mutStateMayLR
)

type state struct {
	mut map[string]int
}

type funcVisitor struct {
}

func (fv *funcVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	switch typed := node.(type) {
	case *ast.GoStmt:
		return nil
	case *ast.IfStmt, *ast.SwitchStmt, *ast.SelectStmt, *ast.TypeSwitchStmt:
		_ = typed
		return nil
	default:
		ast.Walk(&simpleVisitor{}, typed)
		return nil
	}
	return fv
}

type simpleVisitor struct {
}

func (sm *simpleVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	switch typed := node.(type) {
	case *ast.CallExpr:
		for _, arg := range typed.Args {
			ast.Walk(sm, arg)
		}
	case *ast.AssignStmt:
		_ = typed
	}
	return nil
}

type printVisitor struct {
	level int
}

func (fv *printVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	for i := 0; i < fv.level; i++ {
		print("  ")
	}
	println(reflect.TypeOf(node).String())
	switch typed := node.(type) {
	case *ast.GoStmt:
		ast.Walk(&printVisitor{level: fv.level + 1}, typed.Call)
		return nil
	case *ast.SwitchStmt:
		ast.Walk(&printVisitor{level: fv.level + 1}, typed.Body)
		return nil
	case *ast.CallExpr:
		for i := 0; i < fv.level+1; i++ {
			print("  ")
		}
		//println(reflect.TypeOf(typed.Fun).String())
		if sel, ok := typed.Fun.(*ast.SelectorExpr); ok {
			ast.Walk(&printVisitor{level: fv.level + 1}, sel.Sel)
			ast.Walk(&printVisitor{level: fv.level + 1}, sel.X)
		}
		return nil
	case *ast.IfStmt:
		ast.Walk(&printVisitor{level: fv.level + 1}, typed.Body)
		if isif, ok := typed.Else.(*ast.IfStmt); ok {
			ast.Walk(&printVisitor{level: fv.level}, isif)
		} else if typed.Else != nil {
			for i := 0; i < fv.level; i++ {
				print("  ")
			}
			print("else\n")
			ast.Walk(&printVisitor{level: fv.level + 1}, typed.Else)
		}
		return nil
	case *ast.CaseClause:
		for i := 0; i < len(typed.Body); i++ {
			ast.Walk(&printVisitor{level: fv.level + 1}, typed.Body[i])
		}
		return nil
	case *ast.CommClause:
		for i := 0; i < len(typed.Body); i++ {
			ast.Walk(&printVisitor{level: fv.level + 1}, typed.Body[i])
		}
		return nil
	case *ast.SelectStmt:
		ast.Walk(&printVisitor{level: fv.level + 1}, typed.Body)
		return nil
	}
	return fv
}

// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"go/ast"
	"go/token"
)

const (
	exprRead = iota
	exprWrite
	exprExec
)

type stateChanger interface {
	onExpr(op int, obj id, pos token.Pos)
}

type funcVisitor struct {
	sc stateChanger
}

func (fv *funcVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	switch typed := node.(type) {
	case *ast.GoStmt:
		return nil
	case *ast.IfStmt, *ast.SwitchStmt, *ast.SelectStmt, *ast.TypeSwitchStmt:
		return nil
	default:
		sv := &simpleVisitor{sc: fv.sc}
		ast.Walk(sv, typed)
		if sv.handled {
			return nil
		}
	}
	return fv
}

type simpleVisitor struct {
	sc      stateChanger
	handled bool
}

func (sm *simpleVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	sm.handled = true
	switch typed := node.(type) {
	case *ast.CallExpr:
		for _, arg := range typed.Args {
			ast.Walk(sm, arg)
		}
		if expanded := expandSel(typed.Fun); expanded != nil {
			sm.sc.onExpr(exprExec, idFromParts(expanded...), typed.Pos())
		}
	case *ast.AssignStmt:
	case *ast.IncDecStmt:
		//		typed.
	case *ast.BinaryExpr:
	//typed.
	default:
		sm.handled = false
	}
	return nil
}

func expandSel(node ast.Node) []string {
	if sel, ok := node.(*ast.SelectorExpr); ok {
		expanded := expandSel(sel.X)
		if expanded == nil {
			return nil
		}
		return append(expanded, []string{sel.Sel.Name}...)
	} else if id, ok := node.(*ast.Ident); ok {
		return []string{id.Name}
	}
	// TODO(avd) - support for nested CallExpr.
	return nil
}

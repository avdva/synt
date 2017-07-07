// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"go/ast"
	"reflect"
	"strings"
)

const (
	mutStateUnlocked = iota
	mutStateL
	mutStateR
	mutStateMayL
	mutStateMayR
	mutStateMayLR
)

const (
	exprRead = iota
	exprWrite
	exprExec
)

type state struct {
	mut map[string]int
}

type syntChecker struct {
	f   *methodDesc
	pkg *pkgDesc
	st  *state
}

func (sc *syntChecker) onExpr(op int, obj id) {
	println("exec ", obj.String())
	switch op {
	case exprExec:

	}
}

type stateChanger interface {
	onExpr(op int, obj id)
}

type id struct {
	parts []string
}

func idFromParts(parts ...string) id {
	return id{parts: parts}
}

func (i *id) String() string {
	return strings.Join(i.parts, ".")
}

func (i *id) eq(other id) bool {
	if len(i.parts) != len(other.parts) {
		return false
	}
	for i, p := range i.parts {
		if p != other.parts[i] {
			return false
		}
	}
	return true
}

func (i *id) last() string {
	return i.parts[len(i.parts)-1]
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
			sm.sc.onExpr(exprExec, idFromParts(expanded...))
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
	print(reflect.TypeOf(node).String())
	if id, ok := node.(*ast.Ident); ok {
		print(" = ", id.Name)
	}
	println()
	switch typed := node.(type) {
	case *ast.GoStmt:
		ast.Walk(&printVisitor{level: fv.level + 1}, typed.Call)
		return nil
	case *ast.SwitchStmt:
		ast.Walk(&printVisitor{level: fv.level + 1}, typed.Body)
		return nil
	case *ast.CallExpr:
		if sel, ok := typed.Fun.(*ast.SelectorExpr); ok {
			ast.Walk(&printVisitor{level: fv.level + 1}, sel.X)
			ast.Walk(&printVisitor{level: fv.level + 1}, sel.Sel)
		} else {
			ast.Walk(&printVisitor{level: fv.level + 1}, typed.Fun)
		}
		for _, arg := range typed.Args {
			ast.Walk(&printVisitor{level: fv.level + 1}, arg)
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

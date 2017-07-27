// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"go/ast"
	"go/token"
	"strings"
)

const (
	exprRead = iota
	exprWrite
	exprExec
)

type stateChanger interface {
	onExpr(op int, obj id, pos token.Pos)
	onNewContext(node ast.Node)
}

type visitContext struct {
}

type funcVisitor struct {
	vc *visitContext
	sc stateChanger
}

func (fv *funcVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	switch typed := node.(type) {
	case *ast.GoStmt:
		fv.sc.onNewContext(typed.Call)
		return nil
	case *ast.IfStmt, *ast.SwitchStmt, *ast.SelectStmt, *ast.TypeSwitchStmt:
		return nil
	case *ast.CallExpr:
		cv := &callVisitor{sc: fv.sc, parent: fv}
		cid := firstCall(cv.walk(typed))
		if cid.len() > 0 {
			fv.sc.onExpr(exprExec, cid, cv.callPosAt(cid.len()-1))
		}
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

type callVisitor struct {
	sc      stateChanger
	parent  ast.Visitor
	callId  id
	callPos []token.Pos
}

func (cv *callVisitor) walk(node ast.Node) id {
	ast.Walk(cv, node)
	return cv.callId
}

func (cv *callVisitor) callPosAt(i int) token.Pos {
	return cv.callPos[i]
}

func (cv *callVisitor) Visit(node ast.Node) ast.Visitor {
	switch typed := node.(type) {
	case *ast.CallExpr:
		fl, isFuncLit := typed.Fun.(*ast.FuncLit)
		// if it isn't a func literal, we should expand the entire call chain,
		// and then visit arguments.
		if !isFuncLit {
			ast.Walk(cv, typed.Fun)
			n := cv.callId.name()
			cv.callId = cv.callId.selector()
			cv.callId.append(n.String() + "()")
		}
		for _, arg := range typed.Args {
			ast.Walk(cv.parent, arg)
		}
		// if it is a func literal, visit it after visiting all args.
		if isFuncLit {
			ast.Walk(cv.parent, fl)
		}
		return nil
	case *ast.SelectorExpr:
		ast.Walk(cv, typed.X)
		cv.callId.append(typed.Sel.Name)
		cv.callPos = append(cv.callPos, typed.Sel.NamePos)
		return nil
	case *ast.Ident:
		cv.callId.append(typed.Name)
		cv.callPos = append(cv.callPos, typed.NamePos)
		return nil
	}
	return cv
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
		if expanded := expandCall(typed.Fun); expanded != nil {
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

func expandCall(node ast.Node) []string {
	switch typed := node.(type) {
	case *ast.SelectorExpr:
		expanded := expandCall(typed.X)
		if expanded == nil {
			return nil
		}
		return append(expanded, []string{typed.Sel.Name}...)
	case *ast.Ident:
		return []string{typed.Name}
	case *ast.CallExpr:
		expanded := expandCall(typed.Fun)
		if l := len(expanded); l > 0 {
			expanded[l-1] = expanded[l-1] + "()"
		}
		return expanded
	}
	return nil
}

func firstCall(call id) id {
	var result id
	for i := 0; i < call.len(); i++ {
		part := call.part(i)
		if strings.HasSuffix(part, "()") {
			result.append(strings.TrimSuffix(part, "()"))
			break
		}
		result.append(part)
	}
	return result
}

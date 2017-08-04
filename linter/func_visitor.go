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

	exitNormal = 0
	exitReturn = 1
	exitPanic  = 2
)

type stateChanger interface {
	onExpr(op int, obj id, pos token.Pos)
	onNewContext(node ast.Node)
	branchStart(count int) []stateChanger
	branchEnd([]visitResult)
}

type deferItem struct {
	call     ast.Node
	branches [][]deferItem
}

type visitResult struct {
	defers   []deferItem
	exitType int
}

type visitContext struct {
}

type funcVisitor struct {
	vc   *visitContext
	sc   stateChanger
	vr   visitResult
	root bool
}

func newFuncVisitor(sc stateChanger, root bool) *funcVisitor {
	return &funcVisitor{sc: sc, root: root}
}

func (fv *funcVisitor) walk(node ast.Node) visitResult {
	ast.Walk(fv, node)
	if fv.root {
		handleDefers(fv, fv.vr.defers)
	}
	return fv.vr
}

func (fv *funcVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil || fv.vr.exitType != exitNormal {
		return nil
	}
	switch typed := node.(type) {
	case *ast.GoStmt:
		fv.sc.onNewContext(typed.Call)
		return nil
	case *ast.IfStmt:
		fv.handleIf(typed)
		return nil
	case *ast.SwitchStmt, *ast.SelectStmt, *ast.TypeSwitchStmt:
		return nil
	case *ast.DeferStmt:
		for _, arg := range typed.Call.Args {
			ast.Walk(fv, arg)
		}
		fv.vr.defers = append(fv.vr.defers, deferItem{call: typed.Call})
		return nil
	case *ast.CallExpr:
		fv.handleCall(typed)
		return nil
	case *ast.ReturnStmt:
		for _, res := range typed.Results {
			ast.Walk(fv, res)
		}
		fv.vr.exitType = exitReturn
		return nil
	case *ast.AssignStmt:
		if typed.Tok == token.DEFINE {
			//typed.Lhs
		}
		return nil
	}
	return fv
}

func (fv *funcVisitor) handleIf(stmt *ast.IfStmt) {
	var di deferItem
	var results []visitResult
	if stmt.Init != nil {
		ast.Walk(fv, stmt.Init)
	}
	if stmt.Cond != nil {
		ast.Walk(fv, stmt.Cond)
	}
	branches := expandIf(stmt)
	changers := fv.sc.branchStart(len(branches))
	for i, branch := range branches {
		var result visitResult
		fv := newFuncVisitor(changers[i], false)
		for _, node := range branch {
			result = fv.walk(node) // TODO(avd) - is it to correct to leave last result only?
		}
		results = append(results, result)
		di.branches = append(di.branches, result.defers)
	}
	fv.sc.branchEnd(results)
	fv.vr.defers = append(fv.vr.defers, di)
}

func (fv *funcVisitor) handleCall(expr *ast.CallExpr) {
	cv := &callVisitor{sc: fv.sc, parent: fv}
	cid := firstCall(cv.walk(expr))
	if cid.len() > 0 {
		if cid.String() == "panic" {
			fv.vr.exitType = exitPanic
		}
		fv.sc.onExpr(exprExec, cid, cv.callPosAt(cid.len()-1))
	}
}

func handleDefers(fv *funcVisitor, defers []deferItem) {
	for i := len(defers) - 1; i >= 0; i-- {
		di := defers[i]
		if call := di.call; call != nil {
			ast.Walk(fv, call)
			continue
		}
		var results []visitResult
		changers := fv.sc.branchStart(len(di.branches))
		for i, branch := range di.branches {
			fv := newFuncVisitor(changers[i], false)
			handleDefers(fv, branch)
			_, _ = branch, fv
			results = append(results, visitResult{exitType: exitNormal})
		}
		fv.sc.branchEnd(results)
	}
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
			v := newFuncVisitor(cv.sc, true)
			v.walk(fl)
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

func expandIf(node *ast.IfStmt) [][]ast.Node {
	result := [][]ast.Node{[]ast.Node{node.Body}}
	for elseNode := node.Else; elseNode != nil; {
		if ifStmt, ok := elseNode.(*ast.IfStmt); ok {
			var arr []ast.Node
			if ifStmt.Init != nil {
				arr = append(arr, ifStmt.Init)
			}
			if ifStmt.Cond != nil {
				arr = append(arr, ifStmt.Cond)
			}
			arr = append(arr, ifStmt.Body)
			result = append(result, arr)
			elseNode = ifStmt.Else
		} else {
			result = append(result, []ast.Node{elseNode})
			break
		}
	}
	return result
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

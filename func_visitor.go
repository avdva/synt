// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
	"go/token"
	"strings"
)

const (
	exitNormal = 0
	exitReturn = 1
	exitPanic  = 2
)

type stateChanger interface {
	newContext(node ast.Node)
	newObject(obj string, init dotExpr)
	scopeStart()
	scopeEnd()
	branchStart(count int) []stateChanger
	branchEnd([]visitResult)
	expr(op int, obj dotExpr, pos token.Pos)
}

type deferItem struct {
	call     ast.Node
	branches [][]deferItem
}

type visitResult struct {
	defers   []deferItem
	exitType int
}

type visitContext struct{}

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
		fv.sc.newContext(typed.Call)
		return nil
	case *ast.IfStmt:
		fv.handleIf(typed)
		return nil
	case *ast.RangeStmt:
		fv.sc.scopeStart()
		if typed.Tok == token.DEFINE {
			if ident, ok := typed.Key.(*ast.Ident); ok {
				fv.sc.newObject(ident.Name, dotExpr{})
			}
			if ident, ok := typed.Value.(*ast.Ident); ok {
				fv.sc.newObject(ident.Name, dotExpr{})
			}
		}
		ast.Walk(fv, typed.X)
		ast.Walk(fv, typed.Body)
		fv.sc.scopeEnd()
		return nil
	case *ast.ForStmt:
		fv.sc.scopeStart()
		if typed.Init != nil {
			ast.Walk(fv, typed.Init)
		}
		if typed.Cond != nil {
			ast.Walk(fv, typed.Cond)
		}
		if typed.Post != nil {
			ast.Walk(fv, typed.Post)
		}
		ast.Walk(fv, typed.Body)
		fv.sc.scopeEnd()
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
	case *ast.ValueSpec:
		for _, ident := range typed.Names {
			fv.sc.newObject(ident.Name, dotExpr{})
		}
	case *ast.AssignStmt:
		if typed.Tok == token.DEFINE {
			for _, lhs := range typed.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					fv.sc.newObject(ident.Name, dotExpr{})
				}
			}
		}
		//return nil
	}
	return fv
}

func (fv *funcVisitor) handleIf(stmt *ast.IfStmt) {
	var di deferItem
	var results []visitResult
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

func (fv *funcVisitor) handleCall(expr ast.Node) {
	cv := &callVisitor{sc: fv.sc, parent: fv}
	cid := firstCall(cv.walk(expr))
	if cid.len() > 0 {
		if cid.String() == "panic" {
			fv.vr.exitType = exitPanic
		}
		//		fv.sc.expr(opExec, cid, cv.callPosAt(cid.len()-1))
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
			results = append(results, visitResult{exitType: exitNormal})
		}
		fv.sc.branchEnd(results)
	}
}

type callVisitor struct {
	sc      stateChanger
	parent  ast.Visitor
	call    dotExpr
	callPos []token.Pos
}

func (cv *callVisitor) walk(node ast.Node) dotExpr {
	ast.Walk(cv, node)
	return cv.call
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
			n := cv.call.field()
			cv.call = cv.call.selector()
			cv.call.append(n.String() + "()")
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
		cv.call.append(typed.Sel.Name)
		cv.callPos = append(cv.callPos, typed.Sel.NamePos)
		return nil
	case *ast.Ident:
		cv.call.append(typed.Name)
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
			//			sm.sc.expr(opExec, dotExprFromParts(expanded...), typed.Pos())
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
	var arr []ast.Node
	if node.Init != nil {
		arr = append(arr, node.Init)
	}
	if node.Cond != nil {
		arr = append(arr, node.Cond)
	}
	arr = append(arr, node.Body)
	result := [][]ast.Node{arr}
	for elseNode := node.Else; elseNode != nil; {
		if ifStmt, ok := elseNode.(*ast.IfStmt); ok {
			arr = nil
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

func firstCall(call dotExpr) dotExpr {
	var result dotExpr
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

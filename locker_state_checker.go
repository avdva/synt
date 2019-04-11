// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"fmt"
	"go/ast"
	"go/types"
)

const (
	opLock    = "Lock"
	opUnlock  = "Unlock"
	opRLock   = "RLock"
	opRUnlock = "RUnlock"
)

type lockerStateChecker struct {
	types   map[string]struct{}
	lockers map[*ast.Ident]types.Object
}

func newLockerStateChecker(lockTypes []string) *lockerStateChecker {
	checker := &lockerStateChecker{
		types:   make(map[string]struct{}),
		lockers: make(map[*ast.Ident]types.Object),
	}
	for _, typ := range lockTypes {
		checker.types[typ] = struct{}{}
	}
	return checker
}

func (lsc *lockerStateChecker) DoPackage(info *CheckInfo) ([]CheckReport, error) {
	desc, err := makePkgDesc(info.Pkg, info.Fs)
	if err != nil {
		return nil, err
	}
	lsc.collectLockers(desc)
	if len(lsc.lockers) == 0 {
		return nil, nil
	}
	lsc.checkFunctions(info)
	return nil, nil
}

func (lsc *lockerStateChecker) collectLockers(desc *pkgDesc) {
	for ident, obj := range desc.info.Defs {
		if obj == nil {
			continue
		}
		fmt.Printf("%+v %+v\n", ident, obj.Type())
		if _, needed := lsc.types[obj.Type().String()]; needed {
			lsc.lockers[ident] = obj
		}
	}
}

func (lsc *lockerStateChecker) checkFunctions(info *CheckInfo) {
	defs := buildDefs(info.Pkg.Files)
	for name, def := range defs.functions {
		lsc.checkFunction(name, def)
	}
}

func (lsc *lockerStateChecker) checkFunction(name string, def *methodDef) {
	fl := buildFlow(def.node.Body)
	lsc.checkFlow(fl)
}

func (lsc *lockerStateChecker) checkFlow(fl flow) {
	for _, node := range fl {
		if node.branches != nil {
			continue // TODO(avd)
		}
		ops := checkStatements(node.statements)
		_ = ops
	}
}

func checkStatements(statements []ast.Stmt) []op {
	var result []op
	for _, statement := range statements {
		switch typed := statement.(type) {
		case *ast.ExprStmt:
			result = append(result, checkExpr(typed)...)
		}
	}
	return result
}

func checkExpr(expr *ast.ExprStmt) []op {
	var result []op
	switch typed := expr.X.(type) {
	case *ast.CallExpr:
		result = append(result, checkCallExpr(typed)...)
		for _, o := range result {
			fmt.Printf("res %v\n", o.GoString())
		}
	}
	return result
}

func checkCallExpr(expr ast.Expr) []op {
	var result []op
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

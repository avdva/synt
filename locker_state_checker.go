// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
	"go/types"
)

const (
	opLock    = "Lock"
	opUnlock  = "Unlock"
	opRLock   = "RLock"
	opRUnlock = "RUnlock"
)

var (
	opToAct = map[string]lockerAct{
		opLock:    lkActLock,
		opUnlock:  lkActUnlock,
		opRLock:   lkActRLock,
		opRUnlock: lkActRUnlock,
	}
)

type lscConfig struct {
	lockTypes  []string
	autoGuards bool
	filter     func(string) bool
}

type lockerStateChecker struct {
	config  lscConfig
	types   map[string]struct{}
	lockers map[*ast.Ident]types.Object
	ids     *ids
	desc    *pkgDesc
	reports []CheckReport
}

func newLockerStateChecker(config lscConfig) *lockerStateChecker {
	checker := &lockerStateChecker{
		config:  config,
		types:   make(map[string]struct{}),
		lockers: make(map[*ast.Ident]types.Object),
		ids:     newIds(),
	}
	for _, typ := range config.lockTypes {
		checker.types[typ] = struct{}{}
	}
	return checker
}

func (lsc *lockerStateChecker) DoPackage(info *CheckInfo) ([]CheckReport, error) {
	desc, err := makePkgDesc(info.Pkg, info.Fs)
	if err != nil {
		return nil, err
	}
	lsc.desc = desc
	lsc.collectLockers(desc.info)
	if len(lsc.lockers) == 0 {
		return nil, nil
	}
	lsc.checkFunctions(info)
	return lsc.reports, nil
}

func (lsc *lockerStateChecker) collectLockers(info *types.Info) {
	for ident, obj := range info.Defs {
		if obj == nil {
			continue
		}
		if _, needed := lsc.types[objectTypeString(obj)]; needed {
			lsc.lockers[ident] = obj
			lsc.ids.add(obj)
		}
	}
}

func (lsc *lockerStateChecker) idForIdent(id *ast.Ident) string {
	obj := lsc.desc.info.Uses[id]
	if obj == nil {
		return ""
	}
	if _, needed := lsc.types[objectTypeString(obj)]; !needed {
		return ""
	}
	return lsc.ids.strID(obj)
}

func (lsc *lockerStateChecker) checkFunctions(info *CheckInfo) {
	defs := buildDefs(info.Pkg.Files)
	if lsc.config.autoGuards {

	}
	for name, def := range defs.functions {
		if lsc.config.filter == nil || lsc.config.filter(name) {
			lsc.checkFunction(name, def)
		}
	}
}

func (lsc *lockerStateChecker) checkFunction(name string, def *methodDef) {
	fl := buildFlow(def.node.Body)
	states := newLockerStates()
	lsc.checkFlow(fl, states)
	lsc.checkDefers(fl, states)
}

func (lsc *lockerStateChecker) checkFlow(fl flow, states *lockerStates) *lockerStates {
	for _, node := range fl {
		chains := checkStatements(node.statements)
		for _, chain := range chains {
			lsc.checkChain(states, chain)
		}
		if len(node.branches) > 0 {
			var branchStates []*lockerStates
			for _, flow := range node.branches {
				branchStates = append(branchStates, lsc.checkFlow(flow, copyLockerStates(states)))
			}
			states = mergeStates(branchStates)
		}
	}
	return states
}

func (lsc *lockerStateChecker) checkDefers(fl flow, states *lockerStates) *lockerStates {
	for i := len(fl) - 1; i >= 0; i-- {
		node := fl[i]
		chains := checkStatements(node.defers)
		for _, chain := range chains {
			lsc.checkChain(states, chain)
		}
		if len(node.branches) > 0 {
			var branchStates []*lockerStates
			for _, flow := range node.branches {
				branchStates = append(branchStates, lsc.checkDefers(flow, copyLockerStates(states)))
			}
			states = mergeStates(branchStates)
		}
	}
	return states
}

func (lsc *lockerStateChecker) checkChain(states *lockerStates, chain opchain) {
	for i := len(chain) - 1; i >= 0; i-- {
		op := chain[i]
		if op.typ != opExec {
			continue
		}
		act, found := opToAct[op.object.Name]
		if !found || i == 0 {
			continue
		}
		obj := chain[i-1].object // object, on which the operation is performed
		if id := lsc.idForIdent(obj); len(id) > 0 {
			if err := states.stateChange(id, act); err != nil {
				lsc.reports = append(lsc.reports, CheckReport{Err: err, Pos: op.object.Pos()})
			}
		}
	}
}

func checkStatements(statements []ast.Stmt) []opchain {
	var result []opchain
	for _, statement := range statements {
		switch typed := statement.(type) {
		case *ast.ExprStmt:
			result = append(result, checkExpr(typed))
		}
	}
	return result
}

func checkExpr(expr *ast.ExprStmt) opchain {
	var result opchain
	switch typed := expr.X.(type) {
	case *ast.CallExpr:
		// TODO(avd) - check args for non-deffered calls.
		result = append(result, checkCallExpr(typed)...)
	}
	return result
}

func checkCallExpr(expr ast.Expr) opchain {
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

func objectTypeString(obj types.Object) string {
	var strType string
	for obj != nil {
		typ := obj.Type()
		switch typed := typ.(type) {
		case *types.Named:
			strType = typed.String()
			obj = nil
		case *types.Signature:
			if results := typed.Results(); results.Len() == 1 {
				obj = results.At(0)
			} else {
				obj = nil
			}
		case *types.Pointer:
			if named, ok := typed.Elem().(*types.Named); ok {
				strType = named.String()
			}
			obj = nil
		default:
			obj = nil
		}
	}
	return strType
}

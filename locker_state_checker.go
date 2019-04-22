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
	guards  map[*ast.Ident][]*ast.Ident
	ids     *ids
	desc    *pkgDesc
	defs    *defs
	reports []CheckReport
}

func newLockerStateChecker(config lscConfig) *lockerStateChecker {
	checker := &lockerStateChecker{
		config:  config,
		types:   make(map[string]struct{}),
		lockers: make(map[*ast.Ident]types.Object),
		guards:  make(map[*ast.Ident][]*ast.Ident),
		ids:     newIds(),
	}
	for _, typ := range config.lockTypes {
		checker.types[typ] = struct{}{}
	}
	return checker
}

func (lsc *lockerStateChecker) DoPackage(info *CheckInfo) ([]CheckReport, error) {
	if err := lsc.init(info); err != nil {
		return nil, err
	}
	if len(lsc.lockers) == 0 {
		return nil, nil
	}
	lsc.checkFunctions()
	return lsc.reports, nil
}

func (lsc *lockerStateChecker) init(info *CheckInfo) error {
	desc, err := makePkgDesc(info.Pkg, info.Fs)
	if err != nil {
		return err
	}
	lsc.desc = desc
	lsc.collectLockers(desc.info)
	lsc.defs = buildDefs(info.Pkg.Files)
	lsc.buildGuards()
	if lsc.config.autoGuards {

	}
	return nil
}

func (lsc *lockerStateChecker) collectLockers(info *types.Info) {
	for ident, obj := range info.Defs {
		if lsc.lockerTypeNeeded(objectTypeString(obj)) {
			lsc.lockers[ident] = obj
			lsc.ids.add(obj)
		}
	}
}

func (lsc *lockerStateChecker) lockerTypeNeeded(typ string) bool {
	_, needed := lsc.types[typ]
	return needed
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

func (lsc *lockerStateChecker) buildGuards() {
	for _, varDef := range lsc.defs.vars {
		if guards := lsc.guardsFromAnnotations(nil, varDef.annotations); len(guards) > 0 {
			lsc.guards[varDef.node] = guards
		}
	}
	for _, typeDef := range lsc.defs.types {
		for _, fieldDef := range typeDef.fields {
			if guards := lsc.guardsFromAnnotations(typeDef.node.Name, fieldDef.annotations); len(guards) > 0 {
				lsc.guards[fieldDef.node] = guards
			}
		}
	}
}

func (lsc *lockerStateChecker) guardsFromAnnotations(typID *ast.Ident, annotations []annotation) []*ast.Ident {
	var guards []*ast.Ident
	for _, annotation := range annotations {
		for _, g := range annotation.guards() {
			if id := lsc.resolveDotExpr(typID, g.object); id != nil {
				guards = append(guards, id)
			}
		}
	}
	return guards
}

func (lsc *lockerStateChecker) resolveDotExpr(id *ast.Ident, de dotExpr) *ast.Ident {
	if de.len() == 0 {
		return nil
	}
	if mainObj := de.part(0); mainObj != "type" {
		ident, found := lsc.defs.vars[mainObj]
		if !found {
			return nil
		}
		id = ident.node
	}
	for i := 1; i < de.len() && id != nil; i++ {
		def := lsc.desc.info.Defs[id]
		if def == nil {
			return nil
		}
		obj := resolveObjectField(def, de.part(i))
		if obj == nil {
			return nil
		}
		id = lsc.desc.objectsToIdents[obj]
	}
	return id
}

func (lsc *lockerStateChecker) checkFunctions() {
	for name, def := range lsc.defs.functions {
		if lsc.config.filter == nil || lsc.config.filter(name) {
			lsc.checkFunction(name, def)
		}
	}
	for typ, def := range lsc.defs.types {
		for fun, def := range def.methods {
			if lsc.config.filter == nil || lsc.config.filter(typ+"."+fun) {
				lsc.checkFunction(fun, def)
			}
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
	for _, node := range fl.nodes {
		chains := statementsToOpchain(node.statements)
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
	for i := len(fl.nodes) - 1; i >= 0; i-- {
		node := fl.nodes[i]
		chains := statementsToOpchain(node.defers)
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
		case *types.Struct:
			obj = nil
		default:
			obj = nil
		}
	}
	return strType
}

func resolveObjectField(obj types.Object, field string) types.Object {
	typ := obj.Type()
	for {
		if _, ok := typ.(*types.Named); ok {
			typ = typ.Underlying()
		} else {
			break
		}
	}
	st, ok := typ.(*types.Struct)
	if !ok {
		return nil
	}
	for i := 0; i < st.NumFields(); i++ {
		v := st.Field(i)
		if v.Name() == field {
			return v
		}
	}
	return nil
}

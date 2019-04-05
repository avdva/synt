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
	lsc.checkFunctions(desc)
	return nil, nil
}

func (lsc *lockerStateChecker) collectLockers(desc *pkgDesc) {
	for ident, obj := range desc.info.Defs {
		if obj == nil {
			continue
		}
		if _, needed := lsc.types[obj.Type().String()]; needed {
			lsc.lockers[ident] = obj
		}
	}
}

func (lsc *lockerStateChecker) checkFunctions(desc *pkgDesc) {
	for ident, obj := range desc.info.Defs {
		if obj == nil {
			continue
		}
		if obj.Parent().Parent() != types.Universe { // ignore non top-level declarations
			continue
		}
		funcObj, ok := obj.Type().(*types.Signature)
		if !ok {
			continue
		}
		if funcObj.Recv() == nil {
			println(ident.Name)
		}
	}
}

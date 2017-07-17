// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"go/ast"
	"go/token"
	"reflect"
	"strings"

	"github.com/pkg/errors"
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
	mutActLock = iota
	mutActRLock
	mutActUnlock
	mutActRUnlock
)

const (
	exprRead = iota
	exprWrite
	exprExec
)

// stateTable shows how mutex state change in response to mutex actions.
var stateTable = [][]stateChange{
	[]stateChange{ // state is Unlocked
		stateChange{new: mutStateL, err: nil},
		stateChange{new: mutStateR, err: nil},
		stateChange{new: mutStateUnlocked, err: errors.New("unlock of unlocked mutex")},
		stateChange{new: mutStateUnlocked, err: errors.New("unlock of unlocked mutex")},
	},
	[]stateChange{ // state is Locked
		stateChange{new: mutStateL, err: errors.New("lock of locked mutex")},
		stateChange{new: mutStateL, err: errors.New("rlock of locked mutex")},
		stateChange{new: mutStateUnlocked, err: nil},
		stateChange{new: mutStateL, err: errors.New("runlock of locked mutex")},
	},
	[]stateChange{ // state is Rlocked
		stateChange{new: mutStateL, err: errors.New("lock of rlocked mutex")},
		stateChange{new: mutStateR, err: errors.New("rlock of rlocked mutex")},
		stateChange{new: mutStateUnlocked, err: errors.New("unlock of rlocked mutex")},
		stateChange{new: mutStateUnlocked, err: nil},
	},
	[]stateChange{ // state is may be Locked
		stateChange{new: mutStateL, err: errors.New("possible lock of locked mutex")},
		stateChange{new: mutStateMayLR, err: errors.New("possible rlock of locked mutex")},
		stateChange{new: mutStateUnlocked, err: nil},
		stateChange{new: mutStateUnlocked, err: errors.New("possible runlock of locked mutex")},
	},
	[]stateChange{ // state is may be RLocked
		stateChange{new: mutStateL, err: errors.New("possible lock of rlocked mutex")},
		stateChange{new: mutStateMayR, err: errors.New("possible rlock of rlocked mutex")},
		stateChange{new: mutStateUnlocked, err: errors.New("possible unlock of rlocked mutex")},
		stateChange{new: mutStateUnlocked, err: nil},
	},
	[]stateChange{ // state is may be RLocked and Locked
		stateChange{new: mutStateL, err: errors.New("possible lock of locked mutex")},
		stateChange{new: mutStateMayLR, err: errors.New("possible rlock of locked mutex")},
		stateChange{new: mutStateUnlocked, err: errors.New("possible unlock of locked mutex")},
		stateChange{new: mutStateMayL, err: errors.New("possible runlock of locked mutex")},
	},
}

type stateChange struct {
	new int
	err error
}

type syntState struct {
	mut map[string]int
}

func (ss *syntState) stateChange(name string, act int) error {
	old := ss.mut[name]
	new := stateTable[old][act]
	ss.mut[name] = new.new
	return new.err
}

type syntChecker struct {
	pkg       *pkgDesc
	typ       string
	fun       string
	st        *syntState
	currentMD *methodDesc
	reports   []Report
}

func newSyntChecker(pkg *pkgDesc, typ, fun string) *syntChecker {
	result := &syntChecker{
		pkg: pkg,
		typ: typ,
		fun: fun,
		st:  &syntState{mut: make(map[string]int)},
	}
	if len(result.typ) > 0 {
		if typDesc, found := result.pkg.types[result.typ]; found {
			if md, found := typDesc.methods[result.fun]; found {
				result.currentMD = &md
			}
		}
	} else {
		/*if md, found := result.pkg. [result.fun]; found {
			result =  &md
		}*/
	}
	if result.currentMD == nil {
		panic("unknown context")
	}
	return result
}

func (sc *syntChecker) check() {
	ast.Walk(&funcVisitor{sc: sc}, sc.currentMD.node)
}

func (sc *syntChecker) onExpr(op int, obj id, pos token.Pos) {
	println("exec ", obj.String())
	switch op {
	case exprExec:
		errors := sc.onExec(obj)
		for _, e := range errors {
			sc.reports = append(sc.reports, Report{pos: pos, text: e.Error()})
		}
	}
}

func (sc *syntChecker) onExec(obj id) []error {
	var result []error
	sel := obj.selector()
	switch obj.name().String() {
	case "Lock":
		println("        ", obj.String())
		if !sc.currentMD.canCall(obj) {
			result = append(result, errors.Errorf("%s cannot lock %s due to an annotation.", sc.currentMD.name, sel.String()))
		}
		if err := sc.st.stateChange(sel.String(), mutActLock); err != nil {
			result = append(result, err)
		}
	case "RLock":
		if !sc.currentMD.canCall(obj) {
			result = append(result, errors.Errorf("%s cannot rlock %s due to an annotation.", sc.currentMD.name, sel.String()))
		}
		if err := sc.st.stateChange(sel.String(), mutActRLock); err != nil {
			result = append(result, err)
		}
	case "Unlock":
		if err := sc.st.stateChange(sel.String(), mutActUnlock); err != nil {
			result = append(result, err)
		}
	case "RUnlock":
		if err := sc.st.stateChange(sel.String(), mutActRUnlock); err != nil {
			result = append(result, err)
		}
	}
	return result
}

type stateChanger interface {
	onExpr(op int, obj id, pos token.Pos)
}

type id struct {
	parts []string
}

func idFromParts(parts ...string) id {
	return id{parts: parts}
}

func (i id) String() string {
	return strings.Join(i.parts, ".")
}

func (i id) eq(other id) bool {
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

func (i *id) name() id {
	return id{parts: []string{i.parts[len(i.parts)-1]}}
}

func (i id) selector() id {
	return id{parts: i.parts[:len(i.parts)-1]}
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

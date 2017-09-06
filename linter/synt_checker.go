// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/pkg/errors"
)

const mutStateUnknown = -1

const (
	mutStateUnlocked = iota
	mutStateL
	mutStateR
	mutStateMayL
	mutStateMayR
	mutStateMayLR
)

const (
	mutActLock mutexAct = iota
	mutActRLock
	mutActUnlock
	mutActRUnlock
)

var (
	// stateChangeTable shows how mutex state changes in response to mutex actions.
	stateChangeTable = [][]stateChange{
		[]stateChange{ // state is Unlocked
			stateChange{state: mutStateL, err: nil},
			stateChange{state: mutStateR, err: nil},
			stateChange{state: mutStateUnlocked, err: &invalidActError{reason: "not locked"}},
			stateChange{state: mutStateUnlocked, err: &invalidActError{reason: "not locked"}},
		},
		[]stateChange{ // state is Locked
			stateChange{state: mutStateL, err: &invalidActError{reason: "already locked"}},
			stateChange{state: mutStateL, err: &invalidActError{reason: "already locked"}},
			stateChange{state: mutStateUnlocked, err: nil},
			stateChange{state: mutStateL, err: &invalidActError{reason: "locked"}},
		},
		[]stateChange{ // state is Rlocked
			stateChange{state: mutStateL, err: &invalidActError{reason: "already rlocked"}},
			stateChange{state: mutStateR, err: &invalidActError{reason: "already rlocked"}},
			stateChange{state: mutStateUnlocked, err: &invalidActError{reason: "rlocked"}},
			stateChange{state: mutStateUnlocked, err: nil},
		},
		[]stateChange{ // state is may be Locked
			stateChange{state: mutStateL, err: &invalidActError{reason: "already ?locked"}},
			stateChange{state: mutStateMayLR, err: &invalidActError{reason: "already ?locked"}},
			stateChange{state: mutStateUnlocked, err: nil},
			stateChange{state: mutStateUnlocked, err: &invalidActError{reason: "?locked"}},
		},
		[]stateChange{ // state is may be RLocked
			stateChange{state: mutStateL, err: &invalidActError{reason: "already rlocked"}},
			stateChange{state: mutStateMayR, err: &invalidActError{reason: "already rlocked"}},
			stateChange{state: mutStateUnlocked, err: &invalidActError{reason: "?rlocked"}},
			stateChange{state: mutStateUnlocked, err: nil},
		},
		[]stateChange{ // state is may be RLocked and Locked
			stateChange{state: mutStateL, err: &invalidActError{reason: "already ?locked"}},
			stateChange{state: mutStateMayLR, err: &invalidActError{reason: "already ?locked"}},
			stateChange{state: mutStateUnlocked, err: &invalidActError{reason: "?rwlocked"}},
			stateChange{state: mutStateMayL, err: &invalidActError{reason: "?rwlocked"}},
		},
	}
	// stateMergeTable maps two states from branches of code into a result one.
	stateMergeTable = [][]mutexState{
		[]mutexState{ // mutStateUnlocked
			mutStateUnlocked, mutStateMayL, mutStateR, mutStateMayL, mutStateMayR, mutStateMayLR,
		},
		[]mutexState{ // mutStateL
			mutStateMayL, mutStateL, mutStateMayLR, mutStateMayL, mutStateMayLR, mutStateMayLR,
		},
		[]mutexState{ // mutStateR
			mutStateMayR, mutStateMayLR, mutStateR, mutStateMayLR, mutStateMayR, mutStateMayLR,
		},
		[]mutexState{ // mutStateMayL
			mutStateMayL, mutStateMayL, mutStateMayLR, mutStateMayL, mutStateMayLR, mutStateMayLR,
		},
		[]mutexState{ // mutStateMayR
			mutStateMayR, mutStateMayLR, mutStateMayR, mutStateMayLR, mutStateMayR, mutStateMayLR,
		},
		[]mutexState{ // mutStateMayLR
			mutStateMayLR, mutStateMayLR, mutStateMayLR, mutStateMayLR, mutStateMayLR, mutStateMayLR,
		},
	}
)

type mutexState int

func (m mutexState) String() string {
	switch m {
	case mutStateUnlocked:
		return "unlocked"
	case mutStateL:
		return "locked"
	case mutStateR:
		return "rlocked"
	case mutStateMayL:
		return "?locked"
	case mutStateMayR:
		return "?rlocked"
	case mutStateMayLR:
		return "?rwlocked"
	}
	return "unknown"
}

type mutexAct int

func (m mutexAct) String() string {
	switch m {
	case mutActLock:
		return "lock"
	case mutActRLock:
		return "rlock"
	case mutActUnlock:
		return "unlock"
	case mutActRUnlock:
		return "runlock"
	}
	return "unknown"
}

type invalidStateError struct {
	object   string
	expected mutexState
	actual   mutexState
	reason   string
}

func (e invalidStateError) Error() string {
	var pref string
	if len(e.reason) > 0 {
		pref = e.reason + ": "
	}
	return pref + fmt.Sprintf("mutex %q should be %s, but now is %s", e.object, e.expected, e.actual)
}

type invalidActError struct {
	subject string
	object  string
	action  mutexAct
	reason  string
}

func (e invalidActError) Error() string {
	result := fmt.Sprintf("cannot %q %s", e.action, e.object)
	if len(e.subject) > 0 {
		result = e.subject + " " + result
	}
	if len(e.reason) > 0 {
		result = result + " [" + e.reason + "]"
	}
	return result
}

type stateChange struct {
	state mutexState
	err   *invalidActError
}

type syntState struct {
	mut map[string]mutexState
}

func newSyntState() *syntState {
	return &syntState{mut: make(map[string]mutexState)}
}

func (ss *syntState) set(name string, state mutexState) {
	ss.mut[name] = state
}

func (ss *syntState) mutState(name string) (mutexState, bool) {
	state, found := ss.mut[name]
	if !found {
		state = mutStateUnlocked
	}
	return state, found
}

func (ss *syntState) stateChange(id string, act mutexAct) *invalidActError {
	old, _ := ss.mutState(id)
	change := stateChangeTable[old][act]
	ss.mut[id] = change.state
	if change.err == nil {
		return nil
	}
	result := *change.err
	result.action = act
	return &result
}

func (ss *syntState) ensureState(name string, state mutexState) *invalidStateError {
	curState, _ := ss.mutState(name)
	if curState == state {
		return nil
	}
	if state == mutStateR && curState == mutStateL {
		return nil
	}
	return &invalidStateError{object: name, actual: curState, expected: state}
}

type syntChecker struct {
	pkg               *pkgDesc
	typ               string
	fun               string
	branches          []stateChanger
	state             *syntState
	method            *methodDesc
	parsedAnnotations []annotation
	stk               *stack
	ourID             string
	reports           []Report
}

func newSyntChecker(pkg *pkgDesc, typ, fun string) *syntChecker {
	result := &syntChecker{
		pkg:   pkg,
		typ:   typ,
		fun:   fun,
		stk:   newStack(&idGen{}),
		state: newSyntState(),
	}
	result.assignMethod()
	result.buildObjects()
	result.parsedAnnotations = parseAnnotations(result.method.annotations, result.stk)
	return result
}

func (sc *syntChecker) assignMethod() {
	if len(sc.typ) > 0 { // methods
		if typDesc, found := sc.pkg.types[sc.typ]; found {
			if md, found := typDesc.methods[sc.fun]; found {
				sc.method = &md
			}
		}
	} else { // TODO(avd) - functions
	}
	if sc.method == nil {
		panic("unknown context")
	}
}

func (sc *syntChecker) buildObjects() {
	if sc.method.name.len() == 2 {
		sc.stk.push()
		sc.ourID = sc.stk.addObject(sc.method.name.first())
	}
}

func (sc *syntChecker) check() []Report {
	fv := newFuncVisitor(sc, true)
	fv.walk(sc.method.node)
	return sc.reports
}

func (sc *syntChecker) newContext(node ast.Node) {
	newSc := &syntChecker{
		pkg:   sc.pkg,
		typ:   sc.typ,
		state: newSyntState(),
		method: &methodDesc{
			name: sc.method.name,
		},
		stk: copyStack(*sc.stk),
	}
	fv := newFuncVisitor(newSc, true)
	fv.walk(node)
	sc.reports = append(sc.reports, newSc.reports...)
}

func (sc *syntChecker) branchStart(count int) []stateChanger {
	stks := sc.stk.branch(count)
	for i := 0; i < count; i++ {
		newSc := &syntChecker{
			pkg:    sc.pkg,
			typ:    sc.typ,
			state:  copyState(sc.state),
			stk:    stks[i],
			method: sc.method,
		}
		newSc.stk.push()
		sc.branches = append(sc.branches, newSc)
	}
	return sc.branches
}

func (sc *syntChecker) branchEnd(results []visitResult) {
	var states []*syntState
	for i, result := range results {
		bsc := sc.branches[i].(*syntChecker)
		sc.reports = append(sc.reports, bsc.reports...)
		if result.exitType == exitNormal {
			states = append(states, bsc.state)
		}
	}
	if len(states) > 0 {
		sc.state = mergeStates(states)
	}
	sc.branches = nil
}

func (sc *syntChecker) scopeStart() {
	sc.stk.push()
}

func (sc *syntChecker) scopeEnd() {
	sc.stk.pop()
}

func (sc *syntChecker) newObject(name string, init dotExpr) {

	println("----------------")

	for id, obj := range sc.stk.objects {
		println("obj = ", id, "  vars: ")
		for v, k := range obj.vars {
			println("    ", v, " obj: ", k.objectID)
		}
	}
	for id, v := range sc.stk.lastScope().vars {
		println("var ", id, "  of ", v.objectID)
	}
	println("+++++++++++++++++++")

	sc.stk.addObject(dotExprFromParts(name))

	for id, obj := range sc.stk.objects {
		println("obj = ", id, "  vars: ")
		for v, k := range obj.vars {
			println("    ", v, " obj: ", k.objectID)
		}
	}
	for id, v := range sc.stk.lastScope().vars {
		println("var ", id, "  of ", v.objectID)
	}
	println("----------------")
}

func (sc *syntChecker) expr(op int, obj dotExpr, pos token.Pos) {
	switch op {
	case exprExec:
		errors := sc.onExec(obj)
		for _, e := range errors {
			sc.reports = append(sc.reports, Report{pos: pos, err: e})
		}
	}
}

func (sc *syntChecker) onExec(obj dotExpr) []error {
	var result []error
	sel := obj.selector()
	objID := sc.stk.objectIDForExpr(sel)
	if len(objID) == 0 {
		if sel.len() == 1 {
			return []error{errors.Errorf("unknown object: %s", sel.String())}
		}
		objID = sc.stk.addObject(sel)
	}
	switch call := obj.field().String(); call {
	case "Lock", "RLock":
		act := mutActLock
		if call == "RLock" {
			act = mutActRLock
		}
		if !sc.canLock(objID) {
			result = append(result, &invalidActError{
				subject: sc.method.name.field().String(),
				object:  sel.String(),
				action:  act,
				reason:  "annotation",
			})
		}
		if err := sc.state.stateChange(objID, act); err != nil {
			err.object = sel.String()
			result = append(result, err)
		}
	case "Unlock", "RUnlock":
		act := mutActUnlock
		if call == "RUnlock" {
			act = mutActRUnlock
		}
		if err := sc.state.stateChange(objID, act); err != nil {
			err.object = sel.String()
			result = append(result, err)
		}
	default:
		result = sc.checkExec(obj)
	}
	return result
}

func (sc *syntChecker) checkExec(obj dotExpr) []error {
	objID := sc.stk.objectIDForExpr(obj.selector())
	if len(objID) == 0 {
		return nil
	}
	if objID != sc.ourID { // TODO(avd) - call on internal objects.
		return nil
	}
	if len(sc.typ) == 0 { // TODO(avd) - add support for non-member funcs.
		return nil
	}
	calee, found := sc.pkg.types[sc.typ].methods[obj.field().String()]
	if !found {
		return []error{errors.Errorf("unknown method %s", obj.field())}
	}
	var result []error
	caleeStk := newStack(&idGen{})
	caleeStk.push()
	caleeObjID := caleeStk.addObject(calee.name.first())
	parsed := parseAnnotations(calee.annotations, caleeStk)
	_, _ = caleeObjID, parsed
	for _, a := range calee.annotations {
		var state mutexState
		switch a.obj.field().String() {
		case "Lock":
			state = mutStateL
		case "RLock":
			state = mutStateR
		default:
			continue
		}
		if gotLock, err := sc.checkCallerAnnotation(a); err != nil {
			result = append(result, err)
		} else if !gotLock {
			obj := a.obj.copy()
			if obj.first().eq(calee.name.first()) {
				obj.set(0, sc.method.name.part(0))
			}
			id := sc.stk.objectIDForExpr(obj.selector())
			if len(id) == 0 { // TODO(avd) - search by annotation's receiver name.
				continue
			}
			if err := sc.state.ensureState(id, state); err != nil {
				err.reason = "in call to " + calee.name.field().String()
				err.object = obj.selector().String()
				result = append(result, err)
			}
		}
	}
	return result
}

func (sc *syntChecker) checkCallerAnnotation(aCalee annotation) (gotLock bool, err *invalidStateError) {
	caleeName := aCalee.obj.field().String()
	if caleeName != "Lock" && caleeName != "RLock" {
		return
	}
	for _, aCaller := range sc.method.annotations {
		callerName := aCaller.obj.field().String()
		if callerName != "Lock" && callerName != "RLock" {
			continue
		}
		if !aCalee.obj.selector().eq(aCaller.obj.selector()) {
			continue
		}
		if caleeName == "Lock" && callerName == "RLock" {
			err = &invalidStateError{
				object:   aCalee.obj.selector().String(),
				actual:   mutStateR,
				expected: mutStateL,
			}
		} else {
			gotLock = true
		}
		break
	}
	return
}

func (sc *syntChecker) canLock(objID string) bool {
	for _, a := range sc.parsedAnnotations {
		if a.obj.part(0) == objID && !a.not {
			return false
		}
	}
	return true
}

func mergeStates(states []*syntState) *syntState {
	allNames := make(map[string]struct{})
	for _, state := range states {
		for k := range state.mut {
			allNames[k] = struct{}{}
		}
	}
	newState := newSyntState()
	for name, _ := range allNames {
		mutState := mutexState(mutStateUnknown)
		for _, state := range states {
			stateFromBranch, _ := state.mutState(name)
			if mutState == mutStateUnknown {
				mutState = stateFromBranch
				continue
			}
			if mutState == stateFromBranch {
				continue
			}
			mutState = stateMergeTable[mutState][stateFromBranch]
		}
		newState.set(name, mutState)
	}
	return newState
}

func copyState(st *syntState) *syntState {
	result := newSyntState()
	for k, v := range st.mut {
		result.mut[k] = v
	}
	return result
}

func parseAnnotations(annotations []annotation, stk *stack) []annotation {
	var parsedAnnotations []annotation
	for _, a := range annotations {
		sel := a.obj.selector()
		id := stk.objectIDForExpr(sel)
		if len(id) == 0 {
			id = stk.addObject(sel)
		}
		new := annotation{
			not: a.not,
			obj: dotExprFromParts(id, a.obj.field().String()),
		}
		parsedAnnotations = append(parsedAnnotations, new)
	}
	return parsedAnnotations
}

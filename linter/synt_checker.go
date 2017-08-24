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
	mutActLock = iota
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

func (ss *syntState) stateChange(name string, act mutexAct) error {
	old, _ := ss.mutState(name)
	change := stateChangeTable[old][act]
	ss.mut[name] = change.state
	if change.err == nil {
		return nil
	}
	result := *change.err
	result.action = act
	result.object = name
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
	pkg      *pkgDesc
	typ      string
	fun      string
	branches []stateChanger
	state    *syntState
	method   *methodDesc
	stk      *stack
	reports  []Report
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
	if sc.method.id.len() == 2 {
		sc.stk.push()
		sc.stk.addObject(sc.method.id.first())
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
			id: sc.method.id,
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

func (sc *syntChecker) newObject(name string, init id) {

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

	sc.stk.addObject(idFromParts(name))

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

func (sc *syntChecker) expr(op int, obj id, pos token.Pos) {
	switch op {
	case exprExec:
		errors := sc.onExec(obj)
		for _, e := range errors {
			sc.reports = append(sc.reports, Report{pos: pos, err: e})
		}
	}
}

func (sc *syntChecker) onExec(obj id) []error {
	var result []error
	sel := obj.selector()
	switch obj.last().String() {
	case "Lock":
		if !sc.canLock(obj) {
			result = append(result, &invalidActError{
				subject: sc.method.id.last().String(),
				object:  sel.String(),
				action:  mutActLock,
				reason:  "annotation",
			})
		}
		if err := sc.state.stateChange(sel.String(), mutActLock); err != nil {
			result = append(result, err)
		}
	case "RLock":
		if !sc.canLock(obj) {
			result = append(result, &invalidActError{
				subject: sc.method.id.last().String(),
				object:  sel.String(),
				action:  mutActRLock,
				reason:  "annotation",
			})
		}
		if err := sc.state.stateChange(sel.String(), mutActRLock); err != nil {
			result = append(result, err)
		}
	case "Unlock":
		if err := sc.state.stateChange(sel.String(), mutActUnlock); err != nil {
			result = append(result, err)
		}
	case "RUnlock":
		if err := sc.state.stateChange(sel.String(), mutActRUnlock); err != nil {
			result = append(result, err)
		}
	default:
		result = sc.checkExec(obj)
	}
	return result
}

func (sc *syntChecker) checkExec(obj id) []error {
	sel := obj.selector()
	if !sel.eq(sc.method.id.selector()) {
		return nil
	}
	if len(sc.typ) == 0 { // TODO(avd) - add support for non-member funcs.
		return nil
	}
	callee, found := sc.pkg.types[sc.typ].methods[obj.last().String()]
	if !found {
		return []error{errors.Errorf("unknown method %s", obj.last())}
	}
	var result []error
	for _, a := range callee.annotations {
		var state mutexState
		switch a.obj.last().String() {
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
			if err := sc.state.ensureState(a.obj.selector().String(), state); err != nil {
				err.reason = "in call to " + callee.id.last().String()
				result = append(result, err)
			}
		}
	}
	return result
}

func (sc *syntChecker) checkCallerAnnotation(aCalee annotation) (gotLock bool, err *invalidStateError) {
	caleeName := aCalee.obj.last().String()
	if caleeName != "Lock" && caleeName != "RLock" {
		return
	}
	for _, aCaller := range sc.method.annotations {
		callerName := aCaller.obj.last().String()
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

func (sc *syntChecker) canLock(obj id) bool {
	for _, a := range sc.method.annotations {
		if a.obj.selector().eq(obj.selector()) && !a.not {
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

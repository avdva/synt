// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"fmt"
	"go/ast"
	"go/token"

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

// stateTable shows how mutex state change in response to mutex actions.
var stateTable = [][]stateChange{
	[]stateChange{ // state is Unlocked
		stateChange{state: mutStateL, err: nil},
		stateChange{state: mutStateR, err: nil},
		stateChange{state: mutStateUnlocked, err: errors.New("unlock of unlocked mutex")},
		stateChange{state: mutStateUnlocked, err: errors.New("unlock of unlocked mutex")},
	},
	[]stateChange{ // state is Locked
		stateChange{state: mutStateL, err: errors.New("lock of locked mutex")},
		stateChange{state: mutStateL, err: errors.New("rlock of locked mutex")},
		stateChange{state: mutStateUnlocked, err: nil},
		stateChange{state: mutStateL, err: errors.New("runlock of locked mutex")},
	},
	[]stateChange{ // state is Rlocked
		stateChange{state: mutStateL, err: errors.New("lock of rlocked mutex")},
		stateChange{state: mutStateR, err: errors.New("rlock of rlocked mutex")},
		stateChange{state: mutStateUnlocked, err: errors.New("unlock of rlocked mutex")},
		stateChange{state: mutStateUnlocked, err: nil},
	},
	[]stateChange{ // state is may be Locked
		stateChange{state: mutStateL, err: errors.New("possible lock of locked mutex")},
		stateChange{state: mutStateMayLR, err: errors.New("possible rlock of locked mutex")},
		stateChange{state: mutStateUnlocked, err: nil},
		stateChange{state: mutStateUnlocked, err: errors.New("possible runlock of locked mutex")},
	},
	[]stateChange{ // state is may be RLocked
		stateChange{state: mutStateL, err: errors.New("possible lock of rlocked mutex")},
		stateChange{state: mutStateMayR, err: errors.New("possible rlock of rlocked mutex")},
		stateChange{state: mutStateUnlocked, err: errors.New("possible unlock of rlocked mutex")},
		stateChange{state: mutStateUnlocked, err: nil},
	},
	[]stateChange{ // state is may be RLocked and Locked
		stateChange{state: mutStateL, err: errors.New("possible lock of locked mutex")},
		stateChange{state: mutStateMayLR, err: errors.New("possible rlock of locked mutex")},
		stateChange{state: mutStateUnlocked, err: errors.New("possible unlock of locked mutex")},
		stateChange{state: mutStateMayL, err: errors.New("possible runlock of locked mutex")},
	},
}

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
	name             string
	expected, actual mutexState
}

func (e invalidStateError) Error() string {
	return fmt.Sprintf("mutex %q is in state %s, but expected state is %s", e.name, e.actual, e.expected)
}

type invalidActError struct {
	subject, object string
	reason          string
	action          mutexAct
}

func (e invalidActError) Error() string {
	result := fmt.Sprintf("%s cannot %q %s", e.subject, e.action, e.object)
	if len(e.reason) > 0 {
		result = result + " [" + e.reason + "]"
	}
	return result
}

type stateChange struct {
	state mutexState
	err   error
}

type syntState struct {
	mut map[string]mutexState
}

func (ss *syntState) stateChange(name string, act mutexAct) error {
	old := ss.mut[name]
	change := stateTable[old][act]
	ss.mut[name] = change.state
	return change.err
}

func (ss *syntState) ensureState(name string, state mutexState) error {
	curState := ss.mut[name]
	if curState == state {
		return nil
	}
	if state == mutStateR && curState == mutStateL {
		return nil
	}
	return invalidStateError{name: name, actual: curState, expected: state}
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
		st:  &syntState{mut: make(map[string]mutexState)},
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
	switch obj.name().String() {
	case "Lock":
		if !sc.canLock(obj) {
			result = append(result, invalidActError{
				subject: sc.currentMD.obj.name().String(),
				object:  sel.String(),
				action:  mutActLock,
				reason:  "annotation"},
			)
		}
		if err := sc.st.stateChange(sel.String(), mutActLock); err != nil {
			result = append(result, err)
		}
	case "RLock":
		if !sc.canLock(obj) {
			result = append(result, invalidActError{
				subject: sc.currentMD.obj.name().String(),
				object:  sel.String(),
				action:  mutActRLock,
				reason:  "annotation"},
			)
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
	default:
		result = sc.checkExec(obj)
	}
	return result
}

func (sc *syntChecker) checkExec(obj id) []error {
	sel := obj.selector()
	if !sel.eq(sc.currentMD.obj.selector()) {
		return nil
	}
	if len(sc.typ) == 0 { // TODO(avd) - add support for non-member funcs.
		return nil
	}
	callee, found := sc.pkg.types[sc.typ].methods[obj.name().String()]
	if !found {
		return []error{errors.Errorf("unknown method %s", obj.name())}
	}
	var result []error
	for _, a := range callee.annotations {
		var state mutexState
		switch a.obj.name().String() {
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
			if err = sc.st.ensureState(a.obj.selector().String(), state); err != nil {
				result = append(result, err)
			}
		}
	}
	return result
}

func (sc *syntChecker) checkCallerAnnotation(aCalee annotation) (gotLock bool, err error) {
	caleeName := aCalee.obj.name().String()
	if caleeName != "Lock" && caleeName != "RLock" {
		return
	}
	for _, aCaller := range sc.currentMD.annotations {
		callerName := aCaller.obj.name().String()
		if callerName != "Lock" && callerName != "RLock" {
			continue
		}
		if aCalee.obj.selector().eq(aCaller.obj.selector()) {
			if caleeName == "Lock" && callerName == "RLock" {
				err = invalidStateError{name: aCalee.obj.selector().String(), actual: mutStateR, expected: mutStateR}
			} else {
				gotLock = true
			}
			break
		}
	}
	return
}

func (sc *syntChecker) canLock(obj id) bool {
	for _, a := range sc.currentMD.annotations {
		if a.obj.selector().eq(obj.selector()) && !a.not {
			return false
		}
	}
	return true
}

// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import "fmt"

const (
	lkStateUnknown = lockerState(iota)
	lkStateUnlocked
	lkStateL
	lkStateR
	lkStateMayL
	lkStateMayR
	lkStateMayLR
)

const (
	lkActLock = lockerAct(iota)
	lkActRLock
	lkActUnlock
	lkActRUnlock
)

var (
	// stateChangeTable shows how locker state changes in response to lock/unlock actions.
	stateChangeTable = [][]stateChange{
		[]stateChange{ // state is Unlocked
			stateChange{final: lkStateL, err: nil},
			stateChange{final: lkStateR, err: nil},
			stateChange{final: lkStateUnlocked, err: &invalidActError{reason: "not locked"}},
			stateChange{final: lkStateUnlocked, err: &invalidActError{reason: "not locked"}},
		},
		[]stateChange{ // state is Locked
			stateChange{final: lkStateL, err: &invalidActError{reason: "already locked"}},
			stateChange{final: lkStateL, err: &invalidActError{reason: "already locked"}},
			stateChange{final: lkStateUnlocked, err: nil},
			stateChange{final: lkStateL, err: &invalidActError{reason: "locked"}},
		},
		[]stateChange{ // state is Rlocked
			stateChange{final: lkStateL, err: &invalidActError{reason: "already rlocked"}},
			stateChange{final: lkStateR, err: &invalidActError{reason: "already rlocked"}},
			stateChange{final: lkStateUnlocked, err: &invalidActError{reason: "rlocked"}},
			stateChange{final: lkStateUnlocked, err: nil},
		},
		[]stateChange{ // state is may be Locked
			stateChange{final: lkStateL, err: &invalidActError{reason: "already ?locked"}},
			stateChange{final: lkStateMayLR, err: &invalidActError{reason: "already ?locked"}},
			stateChange{final: lkStateUnlocked, err: nil},
			stateChange{final: lkStateUnlocked, err: &invalidActError{reason: "?locked"}},
		},
		[]stateChange{ // state is may be RLocked
			stateChange{final: lkStateL, err: &invalidActError{reason: "already rlocked"}},
			stateChange{final: lkStateMayR, err: &invalidActError{reason: "already rlocked"}},
			stateChange{final: lkStateUnlocked, err: &invalidActError{reason: "?rlocked"}},
			stateChange{final: lkStateUnlocked, err: nil},
		},
		[]stateChange{ // state is may be RLocked and Locked
			stateChange{final: lkStateL, err: &invalidActError{reason: "already ?locked"}},
			stateChange{final: lkStateMayLR, err: &invalidActError{reason: "already ?locked"}},
			stateChange{final: lkStateUnlocked, err: &invalidActError{reason: "?rwlocked"}},
			stateChange{final: lkStateMayL, err: &invalidActError{reason: "?rwlocked"}},
		},
	}
	// stateMergeTable maps two states from branches of code into a result one.
	stateMergeTable = [][]lockerState{
		[]lockerState{ // mutStateUnlocked
			lkStateUnlocked, lkStateMayL, lkStateR, lkStateMayL, lkStateMayR, lkStateMayLR,
		},
		[]lockerState{ // mutStateL
			lkStateMayL, lkStateL, lkStateMayLR, lkStateMayL, lkStateMayLR, lkStateMayLR,
		},
		[]lockerState{ // mutStateR
			lkStateMayR, lkStateMayLR, lkStateR, lkStateMayLR, lkStateMayR, lkStateMayLR,
		},
		[]lockerState{ // mutStateMayL
			lkStateMayL, lkStateMayL, lkStateMayLR, lkStateMayL, lkStateMayLR, lkStateMayLR,
		},
		[]lockerState{ // mutStateMayR
			lkStateMayR, lkStateMayLR, lkStateMayR, lkStateMayLR, lkStateMayR, lkStateMayLR,
		},
		[]lockerState{ // mutStateMayLR
			lkStateMayLR, lkStateMayLR, lkStateMayLR, lkStateMayLR, lkStateMayLR, lkStateMayLR,
		},
	}
)

type stateChange struct {
	final lockerState
	err   *invalidActError
}

type lockerState int

func (ls lockerState) String() string {
	switch ls {
	case lkStateUnlocked:
		return "unlocked"
	case lkStateL:
		return "locked"
	case lkStateR:
		return "rlocked"
	case lkStateMayL:
		return "?locked"
	case lkStateMayR:
		return "?rlocked"
	case lkStateMayLR:
		return "?rwlocked"
	}
	return "unknown"
}

type lockerAct int

func (la lockerAct) String() string {
	switch la {
	case lkActLock:
		return "lock"
	case lkActRLock:
		return "rlock"
	case lkActUnlock:
		return "unlock"
	case lkActRUnlock:
		return "runlock"
	}
	return "unknown"
}

type invalidStateError struct {
	object   string
	expected lockerState
	actual   lockerState
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
	action  lockerAct
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

type lockerStates struct {
	states map[string]lockerState
}

func newLockerStates() *lockerStates {
	return &lockerStates{states: make(map[string]lockerState)}
}

func (ls *lockerStates) set(name string, state lockerState) {
	ls.states[name] = state
}

func (ls *lockerStates) state(name string) (lockerState, bool) {
	state, found := ls.states[name]
	if !found {
		state = lkStateUnlocked
	}
	return state, found
}

func (ls *lockerStates) stateChange(id string, act lockerAct) *invalidActError {
	old, _ := ls.state(id)
	change := stateChangeTable[old][act]
	ls.states[id] = change.final
	if change.err == nil {
		return nil
	}
	result := *change.err
	result.action = act
	return &result
}

func (ls *lockerStates) ensureState(name string, state lockerState) *invalidStateError {
	curState, _ := ls.state(name)
	if curState == state {
		return nil
	}
	if state == lkStateR && curState == lkStateL {
		return nil
	}
	return &invalidStateError{object: name, actual: curState, expected: state}
}

func mergeStates(states []*lockerStates) *lockerStates {
	allNames := make(map[string]struct{})
	for _, state := range states {
		for k := range state.states {
			allNames[k] = struct{}{}
		}
	}
	newState := newLockerStates()
	for name := range allNames {
		resultState := lkStateUnknown
		for _, state := range states {
			objState, _ := state.state(name)
			if resultState == lkStateUnknown {
				resultState = objState
				continue
			}
			if resultState != objState {
				resultState = stateMergeTable[resultState][objState]
			}
		}
		newState.set(name, resultState)
	}
	return newState
}

func copyLockerStates(st *lockerStates) *lockerStates {
	result := newLockerStates()
	for k, v := range st.states {
		result.states[k] = v
	}
	return result
}

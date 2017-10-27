// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import "fmt"

const (
	mutStateUnknown = mutexState(iota)
	mutStateUnlocked
	mutStateL
	mutStateR
	mutStateMayL
	mutStateMayR
	mutStateMayLR
)

const (
	mutActLock = mutexAct(iota)
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

func mergeStates(states []*syntState) *syntState {
	allNames := make(map[string]struct{})
	for _, state := range states {
		for k := range state.mut {
			allNames[k] = struct{}{}
		}
	}
	newState := newSyntState()
	for name := range allNames {
		resultState := mutStateUnknown
		for _, state := range states {
			stateFromBranch, _ := state.mutState(name)
			if resultState == mutStateUnknown {
				resultState = stateFromBranch
				continue
			}
			if resultState != stateFromBranch {
				resultState = stateMergeTable[resultState][stateFromBranch]
			}
		}
		newState.set(name, resultState)
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

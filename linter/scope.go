// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import "strconv"

type object struct {
	id   string
	subs map[string]object
}

type variable struct {
	name     string
	objectID string
}

type scope struct {
	vars map[string]variable
}

type stack struct {
	lastID  int
	objects map[string]object
	scopes  []scope
}

func newScope() scope {
	return scope{
		vars: make(map[string]variable),
	}
}

func newStack() *stack {
	return &stack{
		objects: make(map[string]object),
		scopes:  []scope{newScope()},
	}
}

func (stk *stack) push() {
	println("push")
	stk.scopes = append(stk.scopes, copyScope(*stk.lastScope()))
}

func (stk *stack) pop() {
	println("pop")
	stk.scopes = stk.scopes[:len(stk.scopes)-1]
}

func (stk *stack) lastScope() *scope {
	return &stk.scopes[len(stk.scopes)-1]
}

func (stk *stack) addVar(varName string) {
	id := strconv.Itoa(stk.lastID)
	stk.lastID++
	stk.objects[id] = object{id: id}
	stk.lastScope().vars[varName] = variable{name: varName, objectID: id}
}

func copyScope(sc scope) scope {
	result := newScope()
	for k, v := range sc.vars {
		result.vars[k] = v
	}
	return result
}

func copyStack(stk stack) *stack {
	result := newStack()
	for k, v := range stk.objects {
		result.objects[k] = v
	}
	for _, sc := range stk.scopes {
		stk.scopes = append(stk.scopes, copyScope(sc))
	}
	return result
}

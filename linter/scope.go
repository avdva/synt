// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import "strconv"

type object struct {
	id   string
	vars map[string]variable
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

func (stk *stack) newID() string {
	newID := strconv.Itoa(stk.lastID)
	stk.lastID++
	return newID
}

func (stk *stack) newObject() object {
	return object{id: stk.newID(), vars: make(map[string]variable)}
}

func (stk *stack) addObject(objID id) string {
	println("add ", objID.String())
	if objID.len() == 0 {
		return ""
	}
	if objID.len() == 1 {
		obj := stk.newObject()
		stk.objects[obj.id] = obj
		stk.lastScope().vars[objID.last().String()] = variable{name: objID.last().String(), objectID: obj.id}
		return obj.id
	} else {
		curID := stk.findObjectByVar(objID.part(0))
		if len(curID) == 0 {
			curID = stk.addObject(objID.first())
		}
		obj := stk.objects[curID]
		var lastID string
		for i := 1; i < objID.len(); i++ {
			if v, found := obj.vars[objID.part(i)]; found {
				obj = stk.objects[v.objectID]
			} else {
				newObj := stk.newObject()
				stk.objects[newObj.id] = newObj
				obj.vars[objID.part(i)] = variable{name: objID.part(i), objectID: newObj.id}
				obj = newObj
			}
			lastID = obj.id
		}
		return lastID
	}
}

func (stk *stack) findObjectByVar(varName string) string {
	for i := len(stk.scopes) - 1; i >= 0; i-- {
		sc := stk.scopes[i]
		if v, found := sc.vars[varName]; found {
			return v.objectID
		}
	}
	return ""
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
	for _, sc := range stk.scopes {
		stk.scopes = append(stk.scopes, copyScope(sc))
	}
	return result
}

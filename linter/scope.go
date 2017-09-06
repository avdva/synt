// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import "strconv"

type idGen struct {
	lastID int
}

func (id *idGen) newID() string {
	newID := strconv.Itoa(id.lastID)
	id.lastID++
	return newID
}

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
	gen     *idGen
	objects map[string]object
	scopes  []scope
}

func newScope() scope {
	return scope{
		vars: make(map[string]variable),
	}
}

func newStack(gen *idGen) *stack {
	return &stack{
		gen:     gen,
		objects: make(map[string]object),
		scopes:  []scope{newScope()},
	}
}

func (stk *stack) push() {
	//panic("dsf")
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

func (stk *stack) newObject() object {
	return object{id: stk.gen.newID(), vars: make(map[string]variable)}
}

func (stk *stack) addObject(objID dotExpr) string {
	println("add ", objID.String())
	if objID.len() == 0 {
		return ""
	}
	if objID.len() == 1 {
		obj := stk.newObject()
		stk.objects[obj.id] = obj
		stk.lastScope().vars[objID.field().String()] = variable{name: objID.field().String(), objectID: obj.id}
		return obj.id
	} else {
		curID := stk.objectIDForVar(objID.part(0))
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

func (stk *stack) objectIDForVar(varName string) string {
	for i := len(stk.scopes) - 1; i >= 0; i-- {
		sc := stk.scopes[i]
		if v, found := sc.vars[varName]; found {
			return v.objectID
		}
	}
	return ""
}

func (stk *stack) objectIDForExpr(objID dotExpr) string {
	if objID.len() == 0 {
		return ""
	}
	rootID := stk.objectIDForVar(objID.part(0))
	if objID.len() == 1 || len(rootID) == 0 {
		return rootID
	}
	obj := stk.objects[rootID]
	vars := obj.vars
	for i := 1; i < objID.len(); i++ {
		v, found := vars[objID.part(i)]
		if !found {
			return ""
		}
		obj = stk.objects[v.objectID]
		vars = obj.vars
	}
	return obj.id
}

func (stk *stack) branch(count int) []*stack {
	var result []*stack
	for i := 0; i < count; i++ {
		s := newStack(stk.gen)
		for _, sc := range stk.scopes {
			s.scopes = append(s.scopes, copyScope(sc))
		}
		s.objects = stk.objects
		result = append(result, s)
	}
	return result
}

func copyScope(sc scope) scope {
	result := newScope()
	for k, v := range sc.vars {
		result.vars[k] = v
	}
	return result
}

func copyStack(stk stack) *stack {
	result := newStack(&idGen{lastID: stk.gen.lastID})
	for _, sc := range stk.scopes {
		stk.scopes = append(stk.scopes, copyScope(sc))
	}
	for k, v := range stk.objects {
		stk.objects[k] = v
	}
	return result
}

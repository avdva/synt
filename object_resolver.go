package synt

import (
	"go/types"
	"strconv"
	"strings"
)

type value struct {
	ptr int
}

type object struct {
	obj    types.Object
	values []value
}

type namescope struct {
	symbols map[string]int
}

type objectResolver struct {
	objects []object
	scopes  []namescope
}

func (or *objectResolver) resolve(chain opchain) string {
	if len(chain) == 0 {
		return ""
	}
	first := chain[0]
	if first.Type() != opR {
		return ""
	}
	id := or.findInScope(first.ObjectName())
	if id < 0 {
		return ""
	}
	result := []string{strconv.Itoa(id)}
	for i := 1; i < len(chain); i++ {
		op := chain[i]
		if op.Type() != opR {
			return ""
		}
		_, id = structFieldObjectByName(or.objects[id].obj, op.ObjectName())
		if id < 0 {
			return ""
		}
		result = append(result, strconv.Itoa(id))
	}
	return strings.Join(result, ":")
}

func (or *objectResolver) resolveType(id string) types.Object {
	parts := strings.Split(id, ":")
	if len(parts) == 0 {
		return nil
	}
	intID := atoi(parts[0])
	if intID < 0 || intID >= len(or.objects) {
		return nil
	}
	obj := or.objects[intID].obj
	for i := 1; i < len(parts); i++ {
		intID = atoi(parts[i])
		if intID < 0 {
			return nil
		}
		if obj = structFieldObjectByIndex(obj, intID); obj == nil {
			break
		}
	}
	return obj
}

func (or *objectResolver) addObject(name string, typ types.Object) {
	or.objects = append(or.objects, object{
		obj: typ,
	})
	or.scopes[len(or.scopes)-1].symbols[name] = len(or.objects) - 1
}

func (or *objectResolver) push() {
	or.scopes = append(or.scopes, namescope{
		symbols: make(map[string]int),
	})
}

func (or objectResolver) pop() {
	or.scopes = or.scopes[:len(or.scopes)-1]
}

func (or *objectResolver) findInScope(name string) int {
	for i := len(or.scopes) - 1; i >= 0; i-- {
		if id, found := or.scopes[i].symbols[name]; found {
			return id
		}
	}
	return -1
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}
	return i
}

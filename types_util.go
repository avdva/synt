// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import "go/types"

func objectTypeString(obj types.Object) string {
	var strType string
	for obj != nil {
		typ := obj.Type()
		switch typed := typ.(type) {
		case *types.Named:
			strType = typed.String()
			obj = nil
		case *types.Signature:
			if results := typed.Results(); results.Len() == 1 {
				obj = results.At(0)
			} else {
				obj = nil
			}
		case *types.Pointer:
			if named, ok := typed.Elem().(*types.Named); ok {
				strType = named.String()
			}
			obj = nil
		case *types.Struct:
			obj = nil
		default:
			obj = nil
		}
	}
	return strType
}

func asStructObject(obj types.Object) *types.Struct {
	typ := obj.Type()
	for {
		if _, ok := typ.(*types.Named); ok {
			typ = typ.Underlying()
		} else if ptr, ok := typ.(*types.Pointer); ok {
			typ = ptr.Elem()
		} else {
			break
		}
	}
	st, ok := typ.(*types.Struct)
	if !ok {
		return nil
	}
	return st
}

func structFieldObjectByName(obj types.Object, field string) (types.Object, int) {
	st := asStructObject(obj)
	if st == nil {
		return nil, -1
	}
	for i := 0; i < st.NumFields(); i++ {
		if v := st.Field(i); v.Name() == field {
			return v, i
		}
	}
	return nil, -1
}

func structFieldObjectByIndex(obj types.Object, index int) types.Object {
	st := asStructObject(obj)
	if st == nil {
		return nil
	}
	if index >= st.NumFields() {
		return nil
	}
	return st.Field(index)
}

// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
)

type defsVisitor struct {
	defs *defs
}

func newDefsVisitor() *defsVisitor {
	return &defsVisitor{
		defs: newDefs(),
	}
}

func (sv *defsVisitor) Visit(node ast.Node) ast.Visitor {
	switch typed := node.(type) {
	case *ast.FuncDecl:
		sv.defs.addFuncDecl(typed)
		return nil
	case *ast.TypeSpec:
		sv.defs.addTypeSpec(typed)
		return nil
	case *ast.ValueSpec:
		sv.defs.addValueSpec(typed)
		return nil
	}
	return sv
}

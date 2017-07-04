// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"go/ast"
)

const (
	lockTypeL = iota
	lockTypeR
)

type fileVisitor struct {
	desc *pkgDesc
}

type annotation struct {
	object string
	mutex  string
	lock   int
}

func (fv *fileVisitor) Visit(node ast.Node) ast.Visitor {
	switch typed := node.(type) {
	case *ast.FuncDecl:
		fv.desc.addFuncDecl(typed)
		return nil
	case *ast.TypeSpec:
		fv.desc.addTypeSpec(typed)
		return nil
	}
	return fv
}

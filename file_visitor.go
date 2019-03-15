// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
)

type fileVisitor struct {
	desc *pkgDesc
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

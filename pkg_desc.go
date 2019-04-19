// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
)

type pkgDesc struct {
	info          *types.Info
	typesToIdents map[types.Object]*ast.Ident
}

func makePkgDesc(pkg *ast.Package, fs *token.FileSet) (*pkgDesc, error) {
	var allFiles []*ast.File
	for _, file := range pkg.Files {
		allFiles = append(allFiles, file)
	}
	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Types:  make(map[ast.Expr]types.TypeAndValue),
		Defs:   make(map[*ast.Ident]types.Object),
		Uses:   make(map[*ast.Ident]types.Object),
		Scopes: make(map[ast.Node]*types.Scope),
	}
	if _, err := conf.Check(".", fs, allFiles, info); err != nil {
		return nil, err
	}
	result := &pkgDesc{
		info:          info,
		typesToIdents: make(map[types.Object]*ast.Ident),
	}
	for k, v := range result.info.Defs {
		result.typesToIdents[v] = k
	}
	return result, nil
}

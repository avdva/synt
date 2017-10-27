// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"log"
	"sort"
)

type Linter struct {
	fs  *token.FileSet
	pkg *ast.Package
}

type Report struct {
	pos token.Pos
	err error
}

func (r Report) Error() error {
	return r.err
}

func (r Report) Pos() token.Pos {
	return r.pos
}

func New(fs *token.FileSet, pkg *ast.Package) *Linter {
	return &Linter{fs: fs, pkg: pkg}
}

func (l *Linter) Do() []Report {
	var result []Report
	desc := makePkgDesc(l.pkg, l.fs)
	for typName, typDesc := range desc.types {
		for methodName := range typDesc.methods {
			sc := newSyntChecker(desc, typName, methodName)
			result = append(result, sc.check()...)
		}
	}
	return result
}

func makePkgDesc(pkg *ast.Package, fs *token.FileSet) *pkgDesc {
	var allNames []string
	var allFiles []*ast.File
	for name, file := range pkg.Files {
		allNames = append(allNames, name)
		allFiles = append(allFiles, file)
	}
	sort.Strings(allNames)
	desc := &pkgDesc{
		types:       make(map[string]*typeDesc),
		globalFuncs: make(map[string]*methodDesc),
	}
	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Types:  make(map[ast.Expr]types.TypeAndValue),
		Defs:   make(map[*ast.Ident]types.Object),
		Uses:   make(map[*ast.Ident]types.Object),
		Scopes: make(map[ast.Node]*types.Scope),
	}
	pkga, err := conf.Check(".", fs, allFiles, info)
	if err != nil {
		log.Fatal(err) // type error
	} else {
		_ = pkga
		for i := range info.Defs {
			if i.Obj != nil && i.Obj.Kind == ast.Typ {
				println(i.Name)
			}
		}
		log.Fatal("")
	}

	fv := &fileVisitor{desc}
	for _, name := range allNames {
		file := pkg.Files[name]
		ast.Walk(fv, file)
	}
	return desc
}

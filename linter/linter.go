// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"go/ast"
	"go/token"
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
	desc := makePkgDesc(l.pkg)
	for typName, typDesc := range desc.types {
		for methodName := range typDesc.methods {
			sc := newSyntChecker(desc, typName, methodName)
			result = append(result, sc.check()...)
		}
	}
	return result
}

func makePkgDesc(pkg *ast.Package) *pkgDesc {
	var allFiles []string
	for name := range pkg.Files {
		allFiles = append(allFiles, name)
	}
	sort.Strings(allFiles)
	desc := &pkgDesc{
		types:       make(map[string]*typeDesc),
		globalFuncs: make(map[string]*methodDesc),
	}
	fv := &fileVisitor{desc}
	for _, name := range allFiles {
		file := pkg.Files[name]
		ast.Walk(fv, file)
	}
	return desc
}

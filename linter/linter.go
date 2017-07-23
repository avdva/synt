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

func New(fs *token.FileSet, pkg *ast.Package) *Linter {
	return &Linter{fs: fs, pkg: pkg}
}

func (l *Linter) Do() []Report {
	desc := makePkgDesc(l.pkg)
	_ = desc
	return nil
}

func makePkgDesc(pkg *ast.Package) *pkgDesc {
	var allFiles []string
	for name := range pkg.Files {
		allFiles = append(allFiles, name)
	}
	sort.Strings(allFiles)
	desc := &pkgDesc{
		types:   make(map[string]*typeDesc),
		globals: make(map[string]ast.Node),
	}
	fv := &fileVisitor{desc}
	for _, name := range allFiles {
		file := pkg.Files[name]
		ast.Walk(fv, file)
	}
	return desc
}

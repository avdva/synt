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
}

func New(fs *token.FileSet, pkg *ast.Package) *Linter {
	return &Linter{fs: fs, pkg: pkg}
}

func (l *Linter) makePkgDesc() *pkgDesc {
	var allFiles []string
	for name := range l.pkg.Files {
		allFiles = append(allFiles, name)
	}
	sort.Strings(allFiles)
	desc := &pkgDesc{
		types:   make(map[string]*typeDesc),
		globals: make(map[string]ast.Node),
	}
	fv := &fileVisitor{desc}
	for _, name := range allFiles {
		file := l.pkg.Files[name]
		fv.walk(file)
	}
	return desc
}

func (l *Linter) Do() []Report {
	desc := l.makePkgDesc()
	debugPrintPkgDesc(desc)
	return nil
}

// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"go/ast"
	"go/token"
	"sort"
)

type methodDesc struct {
	node        ast.Node
	annotations []annotation
}

type typeDesc struct {
	methods     map[string]methodDesc
	annotations []annotation
}

type pkgDesc struct {
	types   map[string]typeDesc
	globals map[string]ast.Node
}

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

func (l *Linter) Do() []Report {
	var allFiles []string
	for name := range l.pkg.Files {
		allFiles = append(allFiles, name)
	}
	sort.Strings(allFiles)
	desc := &pkgDesc{
		types:   make(map[string]typeDesc),
		globals: make(map[string]ast.Node),
	}
	for _, name := range allFiles {
		file := l.pkg.Files[name]
		fv := &fileVisitor{desc}
		fv.walk(file)
	}
	return nil
}

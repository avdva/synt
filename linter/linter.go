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

func (l *Linter) Do() []Report {
	var allFiles []string
	for name := range l.pkg.Files {
		allFiles = append(allFiles, name)
	}
	sort.Strings(allFiles)
	for _, name := range allFiles {
		file := l.pkg.Files[name]
		fv := &fileVisitor{}
		fv.walk(file)
	}
	return nil
}

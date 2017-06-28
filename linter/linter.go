// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
)

type methodDesc struct {
	node        ast.Node
	annotations []annotation
}

type fieldDesc struct {
	node        ast.Node
	annotations []annotation
}

type typeDesc struct {
	node    ast.Node
	methods map[string]methodDesc
	fields  map[string]fieldDesc
}

type pkgDesc struct {
	types   map[string]typeDesc
	globals map[string]ast.Node
}

func (d *pkgDesc) addTypeDesc(name string, node ast.Node) {
	td := d.types[name]
	td.node = node
	d.types[name] = td
}

func (d *pkgDesc) addTypeMethod(typName, methodName string, node ast.Node, annotations []annotation) {
	td := d.types[typName]
	if td.methods == nil {
		td.methods = make(map[string]methodDesc)
	}
	td.methods[methodName] = methodDesc{
		node:        node,
		annotations: annotations,
	}
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

func (l *Linter) makePkgDesc() *pkgDesc {
	var allFiles []string
	for name := range l.pkg.Files {
		allFiles = append(allFiles, name)
	}
	sort.Strings(allFiles)
	desc := &pkgDesc{
		types:   make(map[string]typeDesc),
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
	for name, td := range desc.types {
		fmt.Printf("type %s\n", name)
		for _, a := range td.annotations {
			fmt.Printf("  annot = %v\n", a)
		}
		for name, m := range td.methods {
			fmt.Printf("    method %s\n", name)
			for _, a := range m.annotations {
				fmt.Printf("      annot = %v\n", a)
			}
		}
	}
	return nil
}

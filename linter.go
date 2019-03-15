// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
)

type Linter struct {
	fs  *token.FileSet
	pkg *ast.Package
}

type Report struct {
	Err      string
	Location string
}

type reportEntry struct {
	pos token.Pos
	err error
}

func New(fs *token.FileSet, pkg *ast.Package) *Linter {
	return &Linter{fs: fs, pkg: pkg}
}

func (l *Linter) Do() []Report {
	var entries []reportEntry
	desc := makePkgDesc(l.pkg, l.fs)
	for typName, typDesc := range desc.types {
		for methodName := range typDesc.methods {
			sc := newSyntChecker(desc, typName, methodName)
			entries = append(entries, sc.check()...)
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		lhs, rhs := l.fs.Position(entries[i].pos), l.fs.Position(entries[j].pos)
		if lhs.Filename != rhs.Filename {
			return lhs.Filename < rhs.Filename
		}
		return entries[i].pos < entries[j].pos
	})
	var result []Report
	for _, e := range entries {
		result = append(result, Report{Err: e.err.Error(), Location: l.fs.Position(e.pos).String()})
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
		for i, obj := range info.Defs {
			if i.Name == "func3_2" {
				//log.Fatal(obj.Id())
			}
			if i.Obj != nil && i.Obj.Kind == ast.Var && obj != nil {
				println(i.Name, reflect.TypeOf(i.Obj.Decl).String(), fs.Position(i.Pos()).String(), "  : ", obj.Id())
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

func DoDir(name string) ([]Report, error) {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, name, notests, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	for _, pkg := range pkgs {
		return New(fs, pkg).Do(), nil
	}
	return nil, nil
}

func notests(info os.FileInfo) bool {
	return info.IsDir() || !strings.HasSuffix(info.Name(), "_test.go")
}

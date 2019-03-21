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

type checkReport struct {
	pos token.Pos
	err error
}

func New(fs *token.FileSet, pkg *ast.Package) *Linter {
	return &Linter{fs: fs, pkg: pkg}
}

func DoDir(name string) ([]Report, error) {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, name, notests, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var result []Report
	for _, pkg := range pkgs {
		result = append(result, New(fs, pkg).Do()...)
	}
	return result, nil
}

func (l *Linter) Do() []Report {
	reports := checkPackage(l.pkg, l.fs)
	return checkReportsToReports(reports, l.fs)
}

func checkPackage(pkg *ast.Package, fs *token.FileSet) []checkReport {
	var reports []checkReport
	desc := makePkgDesc(pkg, fs)
	for typName, typDesc := range desc.types {
		for methodName := range typDesc.methods {
			sc := newSyntChecker(desc, typName, methodName)
			reports = append(reports, sc.check()...)
		}
	}
	return reports
}

func checkReportsToReports(reports []checkReport, fs *token.FileSet) []Report {
	sort.Slice(reports, func(i, j int) bool {
		lhs, rhs := fs.Position(reports[i].pos), fs.Position(reports[j].pos)
		if lhs.Filename != rhs.Filename {
			return lhs.Filename < rhs.Filename
		}
		return reports[i].pos < reports[j].pos
	})
	var result []Report
	for _, e := range reports {
		result = append(result, Report{
			Err:      e.err.Error(),
			Location: fs.Position(e.pos).String(),
		})
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
	desc := &pkgDesc{
		types:       make(map[string]*typeDesc),
		globalFuncs: make(map[string]*methodDesc),
		info:        info,
	}
	fv := &fileVisitor{desc}
	for _, name := range allNames {
		file := pkg.Files[name]
		ast.Walk(fv, file)
	}
	return desc
}

func notests(info os.FileInfo) bool {
	return info.IsDir() || !strings.HasSuffix(info.Name(), "_test.go")
}

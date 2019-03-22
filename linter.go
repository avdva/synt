// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
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
	var first error
	for _, pkg := range pkgs {
		if reports, err := New(fs, pkg).Do(); err == nil {
			result = append(result, reports...)
		} else if first == nil {
			first = err
		}
	}
	return result, first
}

func (l *Linter) Do() ([]Report, error) {
	reports, err := checkPackage(l.pkg, l.fs)
	if err != nil {
		return nil, err
	}
	return checkReportsToReports(reports, l.fs), nil
}

func checkPackage(pkg *ast.Package, fs *token.FileSet) ([]checkReport, error) {
	var reports []checkReport
	desc, err := makePkgDesc(pkg, fs)
	if err != nil {
		return nil, err
	}
	for typName, typDesc := range desc.types {
		for methodName := range typDesc.methods {
			sc := newSyntChecker(desc, typName, methodName)
			reports = append(reports, sc.check()...)
		}
	}
	return reports, nil
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

func makePkgDesc(pkg *ast.Package, fs *token.FileSet) (*pkgDesc, error) {
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
	_, err := conf.Check(".", fs, allFiles, info)
	if err != nil {
		return nil, err
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
	return desc, nil
}

func notests(info os.FileInfo) bool {
	return info.IsDir() || !strings.HasSuffix(info.Name(), "_test.go")
}

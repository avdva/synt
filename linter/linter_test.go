// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestLinterParseComments(t *testing.T) {
	a := assert.New(t)
	l, err := makeLinter("./test/pkg1", "pkg1")
	if !a.NoError(err) {
		return
	}
	actual := makePkgDesc(l.pkg)
	expected := &pkgDesc{
		types: map[string]*typeDesc{
			"Type1": &typeDesc{
				methods: map[string]methodDesc{
					"func1": methodDesc{
						annotations: []annotation{
							annotation{obj: idFromParts("t", "m", "Lock"), not: true},
						},
					},
					"func2": methodDesc{
						annotations: []annotation{
							annotation{obj: idFromParts("t", "m", "Lock")},
						},
					},
					"func3": methodDesc{
						annotations: []annotation{
							annotation{obj: idFromParts("t", "m", "RLock")},
							annotation{obj: idFromParts("t", "mut", "Lock")},
						},
					},
					"func3_1": methodDesc{},
					"func4":   methodDesc{},
					"func5":   methodDesc{},
					"getM":    methodDesc{},
				},
			},
			"Type2": &typeDesc{
				methods: map[string]methodDesc{},
			},
			"Type3": &typeDesc{
				methods: map[string]methodDesc{},
			},
			"EmptyType": &typeDesc{
				methods: map[string]methodDesc{},
			},
		},
	}
	a.NoError(comparePkgDesc(expected, actual))
}

func TestFunc5(t *testing.T) {
	a := assert.New(t)
	l, err := makeLinter("./test/pkg1", "pkg1")
	if !a.NoError(err) {
		return
	}
	sc, err := makeSyntChecker(l, "Type1", "func5")
	if !a.NoError(err) {
		return
	}
	func5Desc := sc.pkg.types["Type1"].methods["func5"]
	ast.Walk(&printVisitor{w: os.Stdout}, func5Desc.node)
	sc.check()
	for _, rep := range sc.reports {
		println(fmt.Sprintf("%s: %s", rep.text, l.fs.Position(rep.pos).String()))
	}
}

func TestFunc3(t *testing.T) {
	a := assert.New(t)
	l, err := makeLinter("./test/pkg1", "pkg1")
	if !a.NoError(err) {
		return
	}
	sc, err := makeSyntChecker(l, "Type1", "func3")
	if !a.NoError(err) {
		return
	}
	sc.check()
	for _, rep := range sc.reports {
		println(fmt.Sprintf("%s: %s", rep.text, l.fs.Position(rep.pos).String()))
	}
}

func TestFunc3_1(t *testing.T) {
	a := assert.New(t)
	l, err := makeLinter("./test/pkg1", "pkg1")
	if !a.NoError(err) {
		return
	}
	sc, err := makeSyntChecker(l, "Type1", "func3_1")
	if !a.NoError(err) {
		return
	}
	sc.check()
	for _, rep := range sc.reports {
		println(fmt.Sprintf("%s: %s", rep.text, l.fs.Position(rep.pos).String()))
	}
}

func makeLinter(path, pkg string) (*Linter, error) {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	pkgAst, found := pkgs[pkg]
	if !found {
		return nil, errors.Errorf("package %s not found", pkg)
	}
	return New(fs, pkgAst), nil
}

func makeSyntChecker(l *Linter, typ, fun string) (*syntChecker, error) {
	desc := makePkgDesc(l.pkg)
	typeDesc, found := desc.types[typ]
	if !found {
		return nil, errors.Errorf("type %s not found", typ)
	}
	_, found = typeDesc.methods[fun]
	if !found {
		return nil, errors.Errorf("func %s not found", fun)
	}
	sc := newSyntChecker(desc, typ, fun)
	return sc, nil
}

func comparePkgDesc(expected, actual *pkgDesc) error {
	if len(expected.types) != len(actual.types) {
		return errors.Errorf("types count mismatch, expected %d, got %d", len(expected.types), len(actual.types))
	}
	for k, v := range expected.types {
		if td, found := actual.types[k]; !found {
			return errors.Errorf("%s type not found", k)
		} else if err := compareTypeDesc(v, td); err != nil {
			return errors.Wrapf(err, "types %q don't match", k)
		}
	}
	return nil
}

func compareTypeDesc(expected, actual *typeDesc) error {
	if len(expected.methods) != len(actual.methods) {
		return errors.Errorf("methods count mismatch, expected %d, got %d", len(expected.methods), len(actual.methods))
	}
	for k, v := range expected.methods {
		if md, found := actual.methods[k]; !found {
			return errors.Errorf("%s type not found", k)
		} else if err := compareMethodDesc(&v, &md); err != nil {
			return errors.Wrapf(err, "methods %q don't match", k)
		}
	}
	return nil
}

func compareMethodDesc(expected, actual *methodDesc) error {
	if len(expected.annotations) != len(actual.annotations) {
		return errors.Errorf("expected %d, got %d annotations", len(expected.annotations), len(actual.annotations))
	}
	for i, a := range expected.annotations {
		if err := compareAnnotations(a, actual.annotations[i]); err != nil {
			return err
		}
	}
	return nil
}

func compareAnnotations(expected, actual annotation) error {
	if !expected.obj.eq(actual.obj) {
		return errors.Errorf("expected obj %q, got %q", expected.obj.String(), actual.obj.String())
	}
	return nil
}

// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestLinterParseComments(t *testing.T) {
	a := assert.New(t)
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, "./test/pkg1", nil, parser.ParseComments)
	if !a.NoError(err) {
		return
	}
	l := New(fs, pkgs["pkg1"])
	actual := l.makePkgDesc()
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
					"func4": methodDesc{},
					"func5": methodDesc{},
					"getM":  methodDesc{},
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

func TestFuncVisitor(t *testing.T) {
	a := assert.New(t)
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, "./test/pkg1", nil, parser.ParseComments)
	if !a.NoError(err) {
		return
	}
	l := New(fs, pkgs["pkg1"])
	desc := l.makePkgDesc()
	type1Desc, found := desc.types["Type1"]
	if !a.True(found) {
		return
	}
	func5Desc, found := type1Desc.methods["func5"]
	if !a.True(found) {
		return
	}
	ast.Walk(&printVisitor{}, func5Desc.node)
	sc := newSyntChecker(desc, "Type1", "func3")
	ast.Walk(&funcVisitor{sc: sc}, func5Desc.node)
	for _, rep := range sc.reports {
		println(fmt.Sprintf("%s: %s", rep.text, fs.Position(rep.pos)))
	}
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

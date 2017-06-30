// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
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
		types: map[string]typeDesc{
			"Type1": typeDesc{
				methods: map[string]methodDesc{
					"func1": methodDesc{},
					"func2": methodDesc{
						annotations: []annotation{
							annotation{"", "@m", lockTypeL},
						},
					},
					"func3": methodDesc{
						annotations: []annotation{
							annotation{"", "@m", lockTypeL},
						},
					},
					"func4": methodDesc{},
				},
			},
			"Type2": typeDesc{
				methods: map[string]methodDesc{},
			},
			"Type3": typeDesc{
				methods: map[string]methodDesc{},
			},
			"EmptyType": typeDesc{
				methods: map[string]methodDesc{},
			},
		},
	}
	debugPrintPkgDesc(actual)
	a.NoError(comparePkgDesc(expected, actual))
}

func comparePkgDesc(expected, actual *pkgDesc) error {
	if len(expected.types) != len(actual.types) {
		return errors.Errorf("types count mismatch, expected %d, got %d", len(expected.types), len(actual.types))
	}
	for k, v := range expected.types {
		if td, found := actual.types[k]; !found {
			return errors.Errorf("%s type not found", k)
		} else if err := compareTypeDesc(&v, &td); err != nil {
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
	if expected.lock != actual.lock {
		return errors.Errorf("expected lock %d, got %d", expected.lock, actual.lock)
	}
	if expected.mutex != actual.mutex {
		return errors.Errorf("expected mutex %q, got %q", expected.mutex, actual.mutex)
	}
	if expected.object != actual.object {
		return errors.Errorf("expected object %q, got %q", expected.object, actual.object)
	}
	return nil
}

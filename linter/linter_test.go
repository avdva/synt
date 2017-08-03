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
					"func3_2": methodDesc{
						annotations: []annotation{
							annotation{obj: idFromParts("t", "m", "Lock")},
						},
					},
					"func3_3": methodDesc{},
					"func3_4": methodDesc{
						annotations: []annotation{
							annotation{obj: idFromParts("t", "m", "RLock")},
						},
					},
					"func3_5": methodDesc{},
					"func3_6": methodDesc{},
					"func4":   methodDesc{},
					"func5":   methodDesc{},
					"func6":   methodDesc{},
					"func7":   methodDesc{},
					"func8":   methodDesc{},
					"func10":  methodDesc{},
					"func11":  methodDesc{},
					"getM":    methodDesc{},
					"self":    methodDesc{},
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
	sc, err := makeSyntChecker(l.pkg, "Type1", "func9")
	if !a.NoError(err) {
		return
	}
	func5Desc := sc.pkg.types["Type1"].methods["func9"]
	ast.Walk(&printVisitor{w: os.Stdout}, func5Desc.node)
	sc.check()
	for _, rep := range sc.reports {
		println(fmt.Sprintf("%s: %s", rep.err, l.fs.Position(rep.pos).String()))
	}
}

func TestFunc3(t *testing.T) {
	expected := []error{
		&invalidActError{
			subject: "func3",
			object:  "t.m",
			action:  mutActLock,
			reason:  "annotation",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func3")
}

func TestFunc3_1(t *testing.T) {
	expected := []error{
		&invalidStateError{
			object:   "t.m",
			expected: mutStateR,
			actual:   mutStateUnlocked,
			reason:   "in call to func3",
		},
		&invalidStateError{
			object:   "t.mut",
			expected: mutStateL,
			actual:   mutStateUnlocked,
			reason:   "in call to func3",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func3_1")
}

func TestFunc3_3(t *testing.T) {
	expected := []error{
		&invalidStateError{
			object:   "t.m",
			expected: mutStateL,
			actual:   mutStateR,
			reason:   "in call to func3_2",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func3_3")
}

func TestFunc3_4(t *testing.T) {
	expected := []error{
		&invalidStateError{
			object:   "t.m",
			expected: mutStateL,
			actual:   mutStateR,
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func3_4")
}

func TestFunc3_5(t *testing.T) {
	expected := []error{
		&invalidActError{
			subject: "",
			object:  "t.m",
			action:  mutActRUnlock,
			reason:  "not locked",
		},
		&invalidActError{
			subject: "",
			object:  "t.m",
			action:  mutActUnlock,
			reason:  "not locked",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func3_5")
}

func TestFunc3_6(t *testing.T) {
	expected := []error{
		&invalidActError{
			subject: "",
			object:  "t.m",
			action:  mutActUnlock,
			reason:  "not locked",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func3_6")
}

func TestFunc6(t *testing.T) {
	expected := []error{
		&invalidStateError{
			object:   "t.m",
			expected: mutStateL,
			actual:   mutStateUnlocked,
			reason:   "in call to func1",
		},
		&invalidStateError{
			object:   "t.m",
			expected: mutStateL,
			actual:   mutStateUnlocked,
			reason:   "in call to func3_2",
		},
		&invalidStateError{
			object:   "t.m",
			expected: mutStateR,
			actual:   mutStateUnlocked,
			reason:   "in call to func3",
		},
		&invalidStateError{
			object:   "t.mut",
			expected: mutStateL,
			actual:   mutStateUnlocked,
			reason:   "in call to func3",
		},
		&invalidStateError{
			object:   "t.m",
			expected: mutStateR,
			actual:   mutStateUnlocked,
			reason:   "in call to func3",
		},
		&invalidStateError{
			object:   "t.mut",
			expected: mutStateL,
			actual:   mutStateUnlocked,
			reason:   "in call to func3",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func6")
}

func TestFunc7(t *testing.T) {
	expected := []error{
		&invalidStateError{
			object:   "t.mut",
			expected: mutStateL,
			actual:   mutStateUnlocked,
			reason:   "in call to func3",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func7")
}

func TestFunc8(t *testing.T) {
	expected := []error{
		&invalidActError{
			subject: "",
			object:  "t.m",
			action:  mutActLock,
			reason:  "already locked",
		},
		&invalidStateError{
			object:   "t.m",
			expected: 2,
			actual:   mutStateMayLR,
			reason:   "in call to func3_4",
		},
		&invalidActError{
			subject: "",
			object:  "t.m",
			action:  mutActUnlock,
			reason:  "?rwlocked",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func8")
}

func TestFunc9(t *testing.T) {
	expected := []error{
		&invalidActError{subject: "", object: "t.mut", action: 2, reason: "not locked"},
		&invalidStateError{object: "t.mut", expected: 1, actual: 3, reason: "in call to func3"},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func9")
}

func TestFunc10(t *testing.T) {
	expected := []error{
		&invalidStateError{object: "t.m", expected: 1, actual: 3, reason: "in call to func3_2"},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func10")
}

func TestFunc11(t *testing.T) {
	expected := []error{
	//&invalidStateError{object: "t.m", expected: 1, actual: 3, reason: "in call to func3_2"},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func11")
}

func TestFunc12(t *testing.T) {
	expected := []error{
		&invalidActError{subject: "", object: "t.m", action: 0, reason: "already locked"},
		&invalidActError{subject: "", object: "t.m", action: 2, reason: "not locked"},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func12")
}

func TestFunc13(t *testing.T) {
	expected := []error{}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func13")
}

func TestFunc14(t *testing.T) {
	expected := []error{
	//&invalidActError{subject: "", object: "t.m", action: 2, reason: "not locked"},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func14")
}

func TestCandle(t *testing.T) {
	expected := []error{
		&invalidActError{subject: "sourceExists", object: "sd.m", action: 1, reason: "annotation"},
	}
	doTypFuncTest(t, expected, "/home/avd/dev/godev/src/olymptrade.com/olymp-candle-service/processor", "dispatcher", "SourceDispatcher", "sourceExists")
}

func doTypFuncTest(t *testing.T, expected []error, path, pkg, typ, fun string) {
	a := assert.New(t)
	l, err := makeLinter(path, pkg)
	if !a.NoError(err) {
		return
	}
	sc, err := makeSyntChecker(l.pkg, typ, fun)
	if !a.NoError(err) {
		return
	}
	sc.check()
	if !a.Equal(len(expected), len(sc.reports)) {
		return
	}
	for i, rep := range sc.reports {
		a.Equal(expected[i], rep.err)
		println(fmt.Sprintf("%s: %s", rep.err, l.fs.Position(rep.pos).String()))
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

func makeSyntChecker(pkg *ast.Package, typ, fun string) (*syntChecker, error) {
	desc := makePkgDesc(pkg)
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

func compareErrors(a, b error) bool {
	switch typed := a.(type) {
	case invalidActError:
		if ia, ok := b.(invalidActError); ok {
			return typed == ia
		}
	case invalidStateError:
		if is, ok := b.(invalidStateError); ok {
			return typed == is
		}
	}
	return a.Error() == b.Error()
}

// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestLinterParseComments(t *testing.T) {
	a := assert.New(t)
	l, err := New("./test/pkg1", []string{"m"})
	if !a.NoError(err) {
		return
	}
	actual := buildDefs(l.pkgs["pkg1"].Files)
	expected := &defs{
		types: map[string]*typeDef{
			"Type1": &typeDef{
				methods: map[string]methodDef{
					"func1": methodDef{
						annotations: []annotation{
							annotation{obj: dotExprFromParts("t", "m", "Lock"), not: true},
						},
					},
					"func2": methodDef{
						annotations: []annotation{
							annotation{obj: dotExprFromParts("t", "m", "Lock")},
						},
					},
					"func3": methodDef{
						annotations: []annotation{
							annotation{obj: dotExprFromParts("t", "m", "RLock")},
							annotation{obj: dotExprFromParts("t", "mut", "Lock")},
						},
					},
					"func3_1": methodDef{},
					"func3_2": methodDef{
						annotations: []annotation{
							annotation{obj: dotExprFromParts("t", "m", "Lock")},
						},
					},
					"func3_3": methodDef{},
					"func3_4": methodDef{
						annotations: []annotation{
							annotation{obj: dotExprFromParts("t", "m", "RLock")},
						},
					},
					"func3_5": methodDef{},
					"func3_6": methodDef{},
					"func4":   methodDef{},
					"func5":   methodDef{},
					"func6":   methodDef{},
					"func7":   methodDef{},
					"func8":   methodDef{},
					"func10":  methodDef{},
					"func11":  methodDef{},
					"func12":  methodDef{},
					"func13":  methodDef{},
					"func14":  methodDef{},
					"func15":  methodDef{},
					"getM":    methodDef{},
					"self":    methodDef{},
				},
			},
			"Type2": &typeDef{
				methods: map[string]methodDef{},
			},
			"Type3": &typeDef{
				methods: map[string]methodDef{},
			},
			"EmptyType": &typeDef{
				methods: map[string]methodDef{},
			},
		},
	}
	a.NoError(comparePkgDesc(expected, actual))
}

func TestFunc3(t *testing.T) {
	expected := []error{
		&invalidActError{
			subject: "func3",
			object:  "t.m",
			action:  lkActLock,
			reason:  "annotation",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func3")
}

func TestFunc3_1(t *testing.T) {
	expected := []error{
		&invalidStateError{
			object:   "t.m",
			expected: lkStateR,
			actual:   lkStateUnlocked,
			reason:   "in call to func3",
		},
		&invalidStateError{
			object:   "t.mut",
			expected: lkStateL,
			actual:   lkStateUnlocked,
			reason:   "in call to func3",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func3_1")
}

func TestFunc3_3(t *testing.T) {
	expected := []error{
		&invalidStateError{
			object:   "t.m",
			expected: lkStateL,
			actual:   lkStateR,
			reason:   "in call to func3_2",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func3_3")
}

func TestFunc3_3_1(t *testing.T) {
	expected := []error{
		&invalidStateError{
			object:   "t1.m",
			expected: lkStateL,
			actual:   lkStateR,
			reason:   "in call to func3_2",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func3_3_1")
}

func TestFunc3_4(t *testing.T) {
	expected := []error{
		&invalidStateError{
			object:   "t.m",
			expected: lkStateL,
			actual:   lkStateR,
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func3_4")
}

func TestFunc3_5(t *testing.T) {
	expected := []error{
		&invalidActError{
			subject: "",
			object:  "t.m",
			action:  lkActRUnlock,
			reason:  "not locked",
		},
		&invalidActError{
			subject: "",
			object:  "t.m",
			action:  lkActUnlock,
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
			action:  lkActUnlock,
			reason:  "not locked",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func3_6")
}

func TestFunc6(t *testing.T) {
	expected := []error{
		&invalidStateError{
			object:   "t.m",
			expected: lkStateL,
			actual:   lkStateUnlocked,
			reason:   "in call to func1",
		},
		&invalidStateError{
			object:   "t.m",
			expected: lkStateL,
			actual:   lkStateUnlocked,
			reason:   "in call to func3_2",
		},
		&invalidStateError{
			object:   "t.m",
			expected: lkStateR,
			actual:   lkStateUnlocked,
			reason:   "in call to func3",
		},
		&invalidStateError{
			object:   "t.mut",
			expected: lkStateL,
			actual:   lkStateUnlocked,
			reason:   "in call to func3",
		},
		&invalidStateError{
			object:   "t.m",
			expected: lkStateR,
			actual:   lkStateUnlocked,
			reason:   "in call to func3",
		},
		&invalidStateError{
			object:   "t.mut",
			expected: lkStateL,
			actual:   lkStateUnlocked,
			reason:   "in call to func3",
		},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func6")
}

func TestFunc7(t *testing.T) {
	expected := []error{
		&invalidStateError{
			object:   "t.mut",
			expected: lkStateL,
			actual:   lkStateUnlocked,
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
			action:  lkActLock,
			reason:  "already locked",
		},
		&invalidStateError{
			object:   "t.m",
			expected: 2,
			actual:   lkStateMayLR,
			reason:   "in call to func3_4",
		},
		&invalidActError{
			subject: "",
			object:  "t.m",
			action:  lkActUnlock,
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
		&invalidActError{subject: "", object: "t.m", action: 3, reason: "locked"},
		&invalidActError{subject: "", object: "t.m", action: 3, reason: "locked"},
		&invalidActError{subject: "", object: "t.m", action: 3, reason: "locked"},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func14")
}

func TestFunc15(t *testing.T) {
	expected := []error{
		&invalidActError{subject: "", object: "t.m", action: 3, reason: "locked"},
		&invalidActError{subject: "", object: "t.m", action: 3, reason: "locked"},
		&invalidActError{subject: "", object: "t.m", action: 3, reason: "locked"},
		&invalidActError{subject: "", object: "t.m", action: 3, reason: "locked"},
	}
	doTypFuncTest(t, expected, "./test/pkg1", "pkg1", "Type1", "func15")
}

func TestCandle(t *testing.T) {
	expected := []error{
		&invalidActError{subject: "sourceExists", object: "sd.m", action: 1, reason: "annotation"},
	}
	doTypFuncTest(t, expected, "/home/avd/dev/godev/src/olymptrade.com/olymp-candle-service/processor", "dispatcher", "SourceDispatcher", "sourceExists")
}

func doTypFuncTest(t *testing.T, expected []error, path, pkg, typ, fun string) {
	a := assert.New(t)
	l, err := New("./test/pkg1", []string{"m"})
	if !a.NoError(err) {
		return
	}
	sc := newMutexChecker()
	if !a.Equal(len(expected), len(sc.reports)) {
		return
	}
	for i, rep := range sc.reports {
		a.Equal(expected[i], rep.Err)
		println(fmt.Sprintf("%s: %s", rep.Err, l.fs.Position(rep.Pos).String()))
	}
}

func comparePkgDesc(expected, actual *defs) error {
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

func compareTypeDesc(expected, actual *typeDef) error {
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

func compareMethodDesc(expected, actual *methodDef) error {
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

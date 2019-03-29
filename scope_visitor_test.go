// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScopeVisitor(t *testing.T) {
	r := require.New(t)
	l, err := New("./testdata/pkg0", []string{"m"})
	r.NoError(err)
	pkg := l.pkgs["main"]
	r.NotNil(pkg)
	fv := newScopeVisitor()
	for _, file := range pkg.Files {
		ast.Walk(fv, file)
	}
	zeroAstNodes(fv.defs)
	expected := &scopeDefs{
		vars: map[string]*varDef{
			"a": &varDef{},
			"b": &varDef{annotations: []annotation{{obj: dotExprFromParts("m", "Lock")}}},
			"c": &varDef{annotations: []annotation{{obj: dotExprFromParts("m", "Lock")}}},
			"m": &varDef{},
			"n": &varDef{},
		},
		functions: map[string]*methodDef{
			"init":     &methodDef{},
			"main":     &methodDef{},
			"someFunc": &methodDef{},
		},
		types: map[string]*typeDef{},
	}
	r.Equal(expected.vars, fv.defs.vars)
}

func zeroAstNodes(defs *scopeDefs) {
	for _, def := range defs.functions {
		def.node = nil
	}
	for _, def := range defs.types {
		def.expr = nil
	}
	for _, def := range defs.vars {
		def.node = nil
	}
}

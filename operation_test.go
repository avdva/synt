// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOperation1(t *testing.T) {
	r := require.New(t)
	defs := buildDefs(opTestCheckInfo.Pkg.Files)
	fdef := defs.functions["f1"]
	r.NotNil(fdef)
	of := statementsToOpchain(fdef.node.Body.List)
	expected := opFlow{
		opchain{
			op{typ: opRead},
		},
		opchain {
			op{typ: opWrite, object: ast.NewIdent("a")},
		},
	}
	// at this point we're not interested in args.
	zeroArgs(of)
	r.Equal(expected[:2], of[:2])
}

func zeroArgs(flow opFlow) {
	for _, chain := range flow {
		for i := range chain {
			chain[i].args = nil
		}
	}
}

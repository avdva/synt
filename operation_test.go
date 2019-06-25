// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
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
	}
	// at this point we're not interested in args.
	zeroArgs(of)
	r.Equal(expected[:1], of[:1])
}

func zeroArgs(flow opFlow) {
	for _, chain := range flow {
		for i := range chain {
			chain[i].args = nil
		}
	}
}

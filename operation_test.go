// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOperation1(t *testing.T) {
	r := require.New(t)
	defs := buildDefs(opTestCheckInfo.Pkg.Files)
	fdef := defs.functions["f1"]
	r.NotNil(fdef)
	chain := statementsToOpchain(fdef.node.Body.List)
	fmt.Printf("%+v", chain)
}

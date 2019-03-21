// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLinterParsePackage(t *testing.T) {
	r := require.New(t)
	l, err := makeLinter("./testdata/pkg0", "main")
	r.NoError(err)
	desc := makePkgDesc(l.pkg, l.fs)
	for id, object := range desc.info.Defs {
		fmt.Printf("'%q' '%q' '%q' XXX", id.Name, object.Name(), object.Id())
	}
}

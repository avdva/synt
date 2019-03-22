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
	desc, err := makePkgDesc(l.pkg, l.fs)
	r.NoError(err)
	for id, object := range desc.info.Defs {
		if object != nil {
			fmt.Printf("%q %q %q %q\n", id.Name, object.Name(), object.Id(), object.Type())
		}
	}
}

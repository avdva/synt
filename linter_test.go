// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testLocation = "testdata"
	test0Pkg     = "pkg0"
)

func TestLinterParsePackage(t *testing.T) {
	r := require.New(t)
	_, err := New("./testdata/pkg0", []string{"m"})
	r.NoError(err)
}

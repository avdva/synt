// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	pkg0CheckInfo *CheckInfo
)

const (
	testLocation    = "testdata"
	testPkg0Folder  = "pkg0"
	testPkg0Path    = testLocation + "/" + testPkg0Folder
	testPkg0Package = "main"
)

func init() {
	pkgs, fs, err := parsePackage(testPkg0Path)
	if err != nil {
		log.Fatal(err)
	}
	pkg0CheckInfo = &CheckInfo{Pkg: pkgs[testPkg0Package], Fs: fs}
}

func TestLinterParsePackage(t *testing.T) {
	r := require.New(t)
	_, err := New("./testdata/pkg0", []string{"m"})
	r.NoError(err)
}

// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	pkg0CheckInfo *CheckInfo
	pkg1CheckInfo *CheckInfo
)

const (
	testLocation = "testdata"

	testPkg0Folder  = "pkg0"
	testPkg0Path    = testLocation + "/" + testPkg0Folder
	testPkg0Package = "main"

	testPkg1Folder  = "pkg1"
	testPkg1Path    = testLocation + "/" + testPkg1Folder
	testPkg1Package = "pkg1"
)

func init() {
	pkgs, fs, err := parsePackage(testPkg0Path)
	if err != nil {
		log.Fatal(err)
	}
	pkg0CheckInfo = &CheckInfo{Pkg: pkgs[testPkg0Package], Fs: fs}
	pkgs, fs, err = parsePackage(testPkg1Path)
	if err != nil {
		log.Fatal(err)
	}
	pkg1CheckInfo = &CheckInfo{Pkg: pkgs[testPkg1Package], Fs: fs}
}

func TestLinterParsePackage(t *testing.T) {
	r := require.New(t)
	_, err := New("./testdata/pkg0", []string{"m"})
	r.NoError(err)
}

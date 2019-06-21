// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"log"
)

var (
	pkg0CheckInfo   *CheckInfo
	pkg1CheckInfo   *CheckInfo
	opTestCheckInfo *CheckInfo
)

const (
	testLocation = "testdata"

	testPkg0Folder  = "pkg0"
	testPkg0Path    = testLocation + "/" + testPkg0Folder
	testPkg0Package = "main"

	testPkg1Folder  = "pkg1"
	testPkg1Path    = testLocation + "/" + testPkg1Folder
	testPkg1Package = "pkg1"

	testOpFolder  = "optest"
	testOpPath    = testLocation + "/" + testOpFolder
	testOpPackage = "optest"
)

func init() {
	pkgs, fs, err := parsePackage(testPkg0Path)
	if err != nil {
		log.Fatal(err)
	}
	pkg0CheckInfo = &CheckInfo{Pkg: pkgs[testPkg0Package], Fs: fs}

	if pkgs, fs, err = parsePackage(testPkg1Path); err != nil {
		log.Fatal(err)
	}
	pkg1CheckInfo = &CheckInfo{Pkg: pkgs[testPkg1Package], Fs: fs}

	if pkgs, fs, err = parsePackage(testOpPath); err != nil {
		log.Fatal(err)
	}
	opTestCheckInfo = &CheckInfo{Pkg: pkgs[testOpPackage], Fs: fs}
}

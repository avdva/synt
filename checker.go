// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
	"go/token"
)

type CheckInfo struct {
	Pkg *ast.Package
	Fs  *token.FileSet
}

type CheckReport struct {
	Pos token.Pos
	Err error
}

type Checker interface {
	DoPackage(info *CheckInfo) ([]CheckReport, error)
}

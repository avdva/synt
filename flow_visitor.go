// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
)

type flowNode struct {
	statements []ast.Stmt
	branches   []*flowNode
}

type flowVisitor struct {
	nodes []*flowNode
}

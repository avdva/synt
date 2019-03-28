// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"fmt"
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScopeVisitor(t *testing.T) {
	a := assert.New(t)
	l, err := New("./testdata/pkg0", []string{"m"})
	if !a.NoError(err) {
		return
	}
	pkg := l.pkgs["main"]
	a.NotNil(pkg)
	fv := &scopeVisitor{defs: *newScopeDefs()}
	for _, file := range pkg.Files {
		ast.Walk(fv, file)
	}
	fmt.Printf("%+v\n%+v\n%+v\n", fv.defs.vars, fv.defs.functions, fv.defs.types)
	fmt.Printf("%+v", fv.defs.vars["b"].annotations)
}

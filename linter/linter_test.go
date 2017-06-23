// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinterParseComments(t *testing.T) {
	a := assert.New(t)
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, "./test/pkg1", nil, parser.ParseComments)
	if !a.NoError(err) {
		return
	}
	for _, pkg := range pkgs {
		l := New(fs, pkg)
		l.Do()
	}
}

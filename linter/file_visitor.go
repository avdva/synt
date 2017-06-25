// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

const (
	lockTypeL = iota
	lockTypeR
)

type fileVisitor struct {
	desc *pkgDesc
}

type annotation struct {
	object string
	mutex  string
	lock   int
}

func (fv *fileVisitor) walk(node ast.Node) {
	ast.Walk(fv, node)
}

func (fv *fileVisitor) Visit(node ast.Node) ast.Visitor {
	switch typed := node.(type) {
	case *ast.FuncDecl:
		println(reflect.TypeOf(typed.Recv.List[0].Type).String())
		se := typed.Recv.List[0].Type.(*ast.StarExpr)
		println(reflect.TypeOf(se.X).String())
		id := se.X.(*ast.Ident)
		println(id.Name)
		parseComments(typed.Doc)
	case *ast.GenDecl:
		if len(typed.Specs) == 1 {
			if _, ok := typed.Specs[0].(*ast.TypeSpec); ok {
				return &commentGroupVisitor{}
			}
			return nil
		}
	case *ast.TypeSpec:
		parseComments(typed.Doc)
	}
	return fv
}

type commentGroupVisitor struct {
}

func (cgv *commentGroupVisitor) Visit(node ast.Node) ast.Visitor {
	if gr, ok := node.(*ast.CommentGroup); ok {
		parseComments(gr)
	}
	return nil
}

func parseComments(comments *ast.CommentGroup) []annotation {
	const (
		tag = "synt:"
	)
	var result []annotation
	if comments == nil {
		return result
	}
	for _, comment := range comments.List {
		text := strings.Trim(comment.Text, " /*")
		if !strings.HasPrefix(text, tag) {
			continue
		}
		text = text[len(tag):]
		parts := strings.Split(text, ",")
		for _, part := range parts {
			if rec, err := parseRecord(part); err == nil {
				result = append(result, rec)
			}
		}
	}
	fmt.Println(result)
	return result
}

func parseRecord(rec string) (annotation, error) {
	var result annotation
	elems := strings.Split(rec, ":")
	l := len(elems)
	if l < 2 || l > 3 {
		return result, errors.Errorf("invalid directive: %v", rec)
	}
	lockType := elems[l-1]
	if lockType == "L" {
		result.lock = lockTypeL
	} else if lockType == "R" {
		result.lock = lockTypeR
	} else {
		return result, errors.Errorf("invalid lock type: %s", lockType)
	}
	result.mutex = elems[l-2]
	if l == 3 {
		result.object = elems[0]
	}
	return result, nil
}

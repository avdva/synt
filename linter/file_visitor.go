// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"fmt"
	"go/ast"
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
		if typed.Recv == nil {
			return nil
		}
		annotations := parseComments(typed.Doc)
		var typName string
		switch rec := typed.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			id := rec.X.(*ast.Ident)
			typName = id.Name
		case *ast.Ident:
			typName = rec.Name
		}
		fv.desc.addTypeMethod(typName, typed.Name.Name, node, annotations)
	case *ast.GenDecl:
		if len(typed.Specs) == 1 {
			if ts, ok := typed.Specs[0].(*ast.TypeSpec); ok {
				fv.desc.addTypeDesc(ts.Name.Name, ts, parseComments(typed.Doc))
			}
			return nil
		}
	case *ast.TypeSpec:
		if st, ok := typed.Type.(*ast.StructType); ok {
			if typed.Name.Name == "Type3" {
				fmt.Printf("%s %d %v\n", typed.Name.Name, len(st.Fields.List), st.Fields.List[0].Type)
			}
			fv.desc.addTypeDesc(typed.Name.Name, typed, parseComments(typed.Doc))
		}
	}
	return fv
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
	elems := strings.Split(strings.TrimSpace(rec), ":")
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

// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"fmt"
	"go/ast"
	"io"
	"reflect"
	"strings"
)

type printVisitor struct {
	level int
	w     io.Writer
}

func (fv *printVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	fv.w.Write([]byte(strings.Repeat(" ", fv.level*2)))
	fv.w.Write([]byte(reflect.TypeOf(node).String()))
	if id, ok := node.(*ast.Ident); ok {
		fv.w.Write([]byte(" = " + id.Name))
	}
	fv.w.Write([]byte("\n"))
	switch typed := node.(type) {
	case *ast.GoStmt:
		ast.Walk(&printVisitor{level: fv.level + 1, w: fv.w}, typed.Call)
		return nil
	case *ast.SwitchStmt:
		ast.Walk(&printVisitor{level: fv.level + 1, w: fv.w}, typed.Body)
		return nil
	case *ast.CallExpr:
		if sel, ok := typed.Fun.(*ast.SelectorExpr); ok {
			ast.Walk(&printVisitor{level: fv.level + 1, w: fv.w}, sel.X)
			ast.Walk(&printVisitor{level: fv.level + 1, w: fv.w}, sel.Sel)
		} else {
			ast.Walk(&printVisitor{level: fv.level + 1, w: fv.w}, typed.Fun)
		}
		for _, arg := range typed.Args {
			ast.Walk(&printVisitor{level: fv.level + 1, w: fv.w}, arg)
		}
		return nil
	case *ast.IfStmt:
		ast.Walk(&printVisitor{level: fv.level + 1, w: fv.w}, typed.Body)
		if isif, ok := typed.Else.(*ast.IfStmt); ok {
			fv.w.Write([]byte("else "))
			ast.Walk(&printVisitor{level: fv.level, w: fv.w}, isif)
		} else if typed.Else != nil {
			fv.w.Write([]byte(strings.Repeat(" ", fv.level*2)))
			fv.w.Write([]byte("else\n"))
			ast.Walk(&printVisitor{level: fv.level + 1, w: fv.w}, typed.Else)
		}
		return nil
	case *ast.CaseClause:
		for i := 0; i < len(typed.Body); i++ {
			ast.Walk(&printVisitor{level: fv.level + 1, w: fv.w}, typed.Body[i])
		}
		return nil
	case *ast.CommClause:
		for i := 0; i < len(typed.Body); i++ {
			ast.Walk(&printVisitor{level: fv.level + 1, w: fv.w}, typed.Body[i])
		}
		return nil
	case *ast.SelectStmt:
		ast.Walk(&printVisitor{level: fv.level + 1, w: fv.w}, typed.Body)
		return nil
	}
	return fv
}

func debugPrintPkgDesc(desc *pkgDesc) {
	for name, td := range desc.types {
		fmt.Printf("type %s\n", name)
		for name, f := range td.fields {
			fmt.Printf("    field %s\n", name)
			for _, a := range f.annotations {
				fmt.Printf("      annot = %v\n", a)
			}
		}
		for name, m := range td.methods {
			fmt.Printf("    method %s\n", name)
			for _, a := range m.annotations {
				fmt.Printf("      annot = %v\n", a)
			}
		}
	}
}

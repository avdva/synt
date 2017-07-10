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

type annotation struct {
	obj  id
	lock int
	not  bool
}

type methodDesc struct {
	node        ast.Node
	recvName    string
	annotations []annotation
}

func (md *methodDesc) canLock(obj id) bool {
	for _, a := range md.annotations {
		if a.obj.eq(obj) && !a.not {
			return false
		}
	}
	return true
}

type fieldDesc struct {
	node        ast.Node
	annotations []annotation
}

type typeDesc struct {
	node    ast.Node
	methods map[string]methodDesc
	fields  map[string]fieldDesc
}

type pkgDesc struct {
	types   map[string]*typeDesc
	globals map[string]ast.Node
}

func (d *pkgDesc) addFuncDecl(node *ast.FuncDecl) {
	if node.Recv == nil {
		return
	}
	var typName string
	annotations := parseComments(node.Doc)
	recv := node.Recv.List[0]
	switch rec := recv.Type.(type) {
	case *ast.StarExpr:
		id := rec.X.(*ast.Ident)
		typName = id.Name
	case *ast.Ident:
		typName = rec.Name
	}
	td := d.descForType(typName)
	td.methods[node.Name.Name] = methodDesc{
		node:        node,
		recvName:    recv.Names[0].Name,
		annotations: annotations,
	}
}

func (d *pkgDesc) descForType(typName string) *typeDesc {
	td := d.types[typName]
	if td == nil {
		td = &typeDesc{
			methods: make(map[string]methodDesc),
			fields:  make(map[string]fieldDesc),
		}
		d.types[typName] = td
	}
	return td
}

func (d *pkgDesc) addTypeSpec(node *ast.TypeSpec) {
	switch typed := node.Type.(type) {
	case *ast.StructType:
		td := d.descForType(node.Name.Name)
		td.node = node
		if typed.Fields == nil || len(typed.Fields.List) == 0 {
			return
		}
		for _, field := range typed.Fields.List {
			for _, name := range field.Names {
				td.fields[name.Name] = fieldDesc{
					node:        field,
					annotations: parseComments(field.Doc),
				}
			}
		}
	}
}

func parseComments(comments *ast.CommentGroup) []annotation {
	const (
		tag = "synt:"
	)
	var result []annotation
	if comments == nil || len(comments.Text()) == 0 {
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
	return result
}

func parseRecord(rec string) (annotation, error) {
	var result annotation
	elems := strings.Split(strings.TrimSpace(rec), ".")
	l := len(elems)
	if l < 2 {
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
	if strings.HasPrefix(elems[0], "!") {
		result.not = true
		elems[0] = strings.TrimLeft(elems[0], "!")
	}
	result.obj = idFromParts(elems[:len(elems)-1]...)
	return result, nil
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

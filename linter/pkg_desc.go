// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/pkg/errors"
)

type methodDesc struct {
	node        ast.Node
	annotations []annotation
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
	types   map[string]typeDesc
	globals map[string]ast.Node
}

func (d *pkgDesc) addFuncDecl(node *ast.FuncDecl) {
	if node.Recv == nil {
		return
	}
	annotations := parseComments(node.Doc)
	var typName string
	switch rec := node.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		id := rec.X.(*ast.Ident)
		typName = id.Name
	case *ast.Ident:
		typName = rec.Name
	}
	d.addTypeMethod(typName, node.Name.Name, node, annotations)
}

func (d *pkgDesc) addTypeMethod(typName, methodName string, node ast.Node, annotations []annotation) {
	td := d.types[typName]
	if td.methods == nil {
		td.methods = make(map[string]methodDesc)
	}
	td.methods[methodName] = methodDesc{
		node:        node,
		annotations: annotations,
	}
	d.types[typName] = td
}

func (d *pkgDesc) addTypeSpec(node *ast.TypeSpec) {
	if st, ok := node.Type.(*ast.StructType); ok {
		d.addType(node.Name.Name, node)
		if st.Fields != nil && len(st.Fields.List) > 0 {
			for _, f := range st.Fields.List {
				if f.Doc != nil && len(f.Doc.Text()) > 0 {
					for _, name := range f.Names {
						d.addTypeField(node.Name.Name, name.Name, f, parseComments(f.Doc))
					}
				}
			}
		}
	}
}

func (d *pkgDesc) addType(name string, node ast.Node) {
	td := d.types[name]
	td.node = node
	d.types[name] = td
}

func (d *pkgDesc) addTypeField(typName, fieldName string, node ast.Node, annotations []annotation) {
	td := d.types[typName]
	if td.fields == nil {
		td.fields = make(map[string]fieldDesc)
	}
	td.fields[fieldName] = fieldDesc{
		node:        node,
		annotations: annotations,
	}
	d.types[typName] = td
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

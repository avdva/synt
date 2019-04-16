// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
	"strings"
)

type methodDef struct {
	node        *ast.FuncDecl
	annotations []string
}

type varDef struct {
	node        *ast.Ident
	annotations []string
}

type typeDef struct {
	expr    ast.Expr
	methods map[string]methodDef
	fields  map[string]varDef
}

type defs struct {
	types     map[string]*typeDef
	functions map[string]*methodDef
	vars      map[string]*varDef
}

func buildDefs(files map[string]*ast.File) *defs {
	sv := newDefsVisitor()
	for _, file := range files {
		ast.Walk(sv, file)
	}
	return sv.defs
}

func newDefs() *defs {
	return &defs{
		types:     make(map[string]*typeDef),
		functions: make(map[string]*methodDef),
		vars:      make(map[string]*varDef),
	}
}

func (d *defs) addFuncDecl(node *ast.FuncDecl) {
	if node.Recv == nil {
		d.addFunc(node)
	} else {
		d.addMethod(node)
	}
}

func (d *defs) addFunc(node *ast.FuncDecl) {
	annotations := parseComments(node.Doc)
	d.functions[node.Name.Name] = &methodDef{
		node:        node,
		annotations: annotations,
	}
}

func (d *defs) addMethod(node *ast.FuncDecl) {
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
	td.methods[node.Name.Name] = methodDef{
		node:        node,
		annotations: annotations,
	}
}

func (d *defs) descForType(typName string) *typeDef {
	td := d.types[typName]
	if td == nil {
		td = &typeDef{
			methods: make(map[string]methodDef),
			fields:  make(map[string]varDef),
		}
		d.types[typName] = td
	}
	return td
}

func (d *defs) addTypeSpec(node *ast.TypeSpec) {
	td := d.descForType(node.Name.Name)
	switch typed := node.Type.(type) {
	case *ast.StructType:
		if typed.Fields == nil || len(typed.Fields.List) == 0 {
			return
		}
		for _, field := range typed.Fields.List {
			for _, name := range field.Names {
				td.fields[name.Name] = varDef{
					node:        name,
					annotations: parseComments(field.Doc),
				}
			}
		}
	}
}

func (d *defs) addValueSpec(node *ast.ValueSpec) {
	for _, name := range node.Names {
		d.vars[name.Name] = &varDef{
			node:        name,
			annotations: parseComments(node.Doc),
		}
	}
}

func parseComments(comments *ast.CommentGroup) []string {
	const (
		tag = "synt:"
	)
	var result []string
	if comments == nil || len(comments.Text()) == 0 {
		return result
	}
	for _, comment := range comments.List {
		text := strings.Trim(comment.Text, " /*")
		if !strings.HasPrefix(text, tag) {
			continue
		}
		text = text[len(tag):]
		parts := strings.Split(text, " ")
		for _, part := range parts {
			result = append(result, part)
		}
	}
	return result
}

type annotation struct {
	obj dotExpr
	not bool
}

func parseRecord(rec string) annotation {
	var result annotation
	rec = strings.TrimSpace(rec)
	if strings.HasPrefix(rec, "!") {
		result.not = true
		rec = strings.TrimLeft(rec, "!")
	}
	result.obj = dotExprFromParts(strings.Split(rec, ".")...)
	return result
}

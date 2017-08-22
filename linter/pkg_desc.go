// Copyright 2017 Aleksandr Demakin. All rights reserved.

package linter

import (
	"go/ast"
	"strings"
)

type id struct {
	parts []string
}

func idFromParts(parts ...string) id {
	return id{parts: parts}
}

func (i id) String() string {
	return strings.Join(i.parts, ".")
}

func (i id) len() int {
	return len(i.parts)
}

func (i id) part(idx int) string {
	return i.parts[idx]
}

func (i id) eq(other id) bool {
	if len(i.parts) != len(other.parts) {
		return false
	}
	for i, p := range i.parts {
		if p != other.parts[i] {
			return false
		}
	}
	return true
}

func (i *id) last() id {
	return id{parts: []string{i.parts[len(i.parts)-1]}}
}

func (i id) selector() id {
	return id{parts: i.parts[:len(i.parts)-1]}
}

func (i id) first() id {
	return id{parts: []string{i.parts[0]}}
}

func (i *id) append(part string) {
	i.parts = append(i.parts, part)
}

type annotation struct {
	obj id
	not bool
}

type methodDesc struct {
	node        ast.Node
	id          id
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
	types       map[string]*typeDesc
	globalFuncs map[string]*methodDesc
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
		id:          idFromParts(recv.Names[0].Name, node.Name.Name),
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
			result = append(result, parseRecord(part))
		}
	}
	return result
}

func parseRecord(rec string) annotation {
	var result annotation
	rec = strings.TrimSpace(rec)
	if strings.HasPrefix(rec, "!") {
		result.not = true
		rec = strings.TrimLeft(rec, "!")
	}
	result.obj = idFromParts(strings.Split(rec, ".")...)
	return result
}

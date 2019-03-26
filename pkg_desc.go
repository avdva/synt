// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"sort"
	"strings"
)

type dotExpr struct {
	parts []string
}

func dotExprFromParts(parts ...string) dotExpr {
	return dotExpr{parts: parts}
}

func (i dotExpr) String() string {
	return strings.Join(i.parts, ".")
}

func (i dotExpr) len() int {
	return len(i.parts)
}

func (i dotExpr) part(idx int) string {
	return i.parts[idx]
}

func (i dotExpr) eq(other dotExpr) bool {
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

func (i *dotExpr) field() dotExpr {
	return dotExpr{parts: []string{i.parts[len(i.parts)-1]}}
}

func (i dotExpr) selector() dotExpr {
	return dotExpr{parts: i.parts[:len(i.parts)-1]}
}

func (i dotExpr) first() dotExpr {
	return dotExpr{parts: []string{i.parts[0]}}
}

func (i dotExpr) copy() dotExpr {
	return dotExprFromParts(i.parts...)
}

func (i dotExpr) set(idx int, part string) {
	i.parts[idx] = part
}

func (i *dotExpr) append(part string) {
	i.parts = append(i.parts, part)
}

type annotation struct {
	obj dotExpr
	not bool
}

type methodDesc struct {
	node        ast.Node
	name        dotExpr
	annotations []annotation
}

type fieldDesc struct {
	node        ast.Node
	annotations []annotation
}

type typeDesc struct {
	expr    ast.Expr
	methods map[string]methodDesc
	fields  map[string]fieldDesc
}

type pkgDesc struct {
	info        *types.Info
	types       map[string]*typeDesc
	globalFuncs map[string]*methodDesc
}

func makePkgDesc(pkg *ast.Package, fs *token.FileSet) (*pkgDesc, error) {
	var allNames []string
	var allFiles []*ast.File
	for name, file := range pkg.Files {
		allNames = append(allNames, name)
		allFiles = append(allFiles, file)
	}
	sort.Strings(allNames)
	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Types:  make(map[ast.Expr]types.TypeAndValue),
		Defs:   make(map[*ast.Ident]types.Object),
		Uses:   make(map[*ast.Ident]types.Object),
		Scopes: make(map[ast.Node]*types.Scope),
	}
	_, err := conf.Check(".", fs, allFiles, info)
	if err != nil {
		return nil, err
	}
	desc := &pkgDesc{
		types:       make(map[string]*typeDesc),
		globalFuncs: make(map[string]*methodDesc),
		info:        info,
	}
	fv := &fileVisitor{desc}
	for _, name := range allNames {
		file := pkg.Files[name]
		ast.Walk(fv, file)
	}
	return desc, nil
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
		name:        dotExprFromParts(recv.Names[0].Name, node.Name.Name),
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
		//		td.expr = node
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
	result.obj = dotExprFromParts(strings.Split(rec, ".")...)
	return result
}

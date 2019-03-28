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

type annotation struct {
	obj dotExpr
	not bool
}

type methodDef struct {
	node        ast.Node
	name        dotExpr
	annotations []annotation
}

type varDef struct {
	node        ast.Node
	annotations []annotation
}

type typeDef struct {
	expr    ast.Expr
	methods map[string]methodDef
	fields  map[string]varDef
}

type scopeDefs struct {
	types     map[string]*typeDef
	functions map[string]*methodDef
	vars      map[string]*varDef
}

func newScopeDefs() *scopeDefs {
	return &scopeDefs{
		types:     make(map[string]*typeDef),
		functions: make(map[string]*methodDef),
		vars:      make(map[string]*varDef),
	}
}

func makePkgDesc(pkg *ast.Package, fs *token.FileSet) (*scopeDefs, error) {
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
	desc := &scopeDefs{
		types:     make(map[string]*typeDef),
		functions: make(map[string]*methodDef),
	}
	fv := &scopeVisitor{}
	for _, name := range allNames {
		file := pkg.Files[name]
		ast.Walk(fv, file)
	}
	return desc, nil
}

func (d *scopeDefs) addFuncDecl(node *ast.FuncDecl) {
	if node.Recv == nil {
		d.addFunc(node)
	} else {
		d.addMethod(node)
	}
}

func (d *scopeDefs) addFunc(node *ast.FuncDecl) {
	annotations := parseComments(node.Doc)
	d.functions[node.Name.Name] = &methodDef{
		node:        node,
		name:        dotExprFromParts(node.Name.Name),
		annotations: annotations,
	}
}

func (d *scopeDefs) addMethod(node *ast.FuncDecl) {
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
		name:        dotExprFromParts(recv.Names[0].Name, node.Name.Name),
		annotations: annotations,
	}
}

func (d *scopeDefs) descForType(typName string) *typeDef {
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

func (d *scopeDefs) addTypeSpec(node *ast.TypeSpec) {
	switch typed := node.Type.(type) {
	case *ast.StructType:
		td := d.descForType(node.Name.Name)
		if typed.Fields == nil || len(typed.Fields.List) == 0 {
			return
		}
		for _, field := range typed.Fields.List {
			for _, name := range field.Names {
				td.fields[name.Name] = varDef{
					node:        field,
					annotations: parseComments(field.Doc),
				}
			}
		}
	}
}

func (d *scopeDefs) addValueSpec(node *ast.ValueSpec) {
	for _, name := range node.Names {
		d.vars[name.Name] = &varDef{
			node:        name,
			annotations: parseComments(node.Doc),
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

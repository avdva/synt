// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
)

const (
	opR opType = iota
	opW
	opExec
)

type opType int

func (t opType) String() string {
	switch t {
	case opR:
		return "r"
	case opW:
		return "w"
	case opExec:
		return "e"
	default:
		return "unknown"
	}
}

type operation interface {
	Type() opType
	ObjectName() string
	String() string
	Pos() token.Pos
}

type opFlow []opchain

func (of opFlow) String() string {
	var buff bytes.Buffer
	for i, o := range of {
		if i != 0 {
			buff.WriteString("→")
		}
		if o == nil {
			buff.WriteString("<nil>")
		} else {
			buff.WriteString(o.String())
		}
	}
	return buff.String()
}

func statementsToOpchain(statements []ast.Stmt, idents map[*ast.Ident]types.Object) opFlow {
	var result opFlow
	for _, statement := range statements {
		result = append(result, statementToOpchain(statement, idents)...)
	}
	return result
}

func statementToOpchain(statement ast.Stmt, idents map[*ast.Ident]types.Object) opFlow {
	var result opFlow
	switch typed := statement.(type) {
	case *ast.AssignStmt:
		result = append(result, expandAssignment(typed.Lhs, typed.Rhs, idents)...)
	case *ast.ExprStmt:
		result = append(result, expandExprStmt(typed, idents)...)
	case *ast.IncDecStmt:
		result = append(result, expandIncDec(typed, idents)...)
	case nil:
	default:
		fmt.Printf("statementToOpchain: skipping %s\n", reflect.ValueOf(statement).Type().String())
	}
	return result
}

func expandIncDec(stmt *ast.IncDecStmt, idents map[*ast.Ident]types.Object) opFlow {
	var result opFlow
	expanded := expandExpr(stmt.X, idents)
	if len(expanded) > 1 {
		result = append(result, expanded[:len(expanded)-1])
	}
	return result
}

func expandAssignment(lhs, rhs []ast.Expr, idents map[*ast.Ident]types.Object) opFlow {
	var result opFlow
	var writes []wOp
	for _, lhs := range lhs {
		expanded := expandExpr(lhs, idents)
		if fl := flatten(expanded, 1); len(fl) > 1 {
			result = append(result, fl[:len(fl)-1])
		}
		writes = append(writes, wOp{lhs: expanded})
	}
	for i, rhs := range rhs {
		expanded := expandExpr(rhs, idents)
		result = append(result, expanded)
		writes[i].rhs = expanded
	}
	for i := range writes {
		result = append(result, []operation{&writes[i]})
	}
	return result
}

func expandExprStmt(expr *ast.ExprStmt, idents map[*ast.Ident]types.Object) opFlow {
	var result opFlow
	switch typed := expr.X.(type) {
	case *ast.CallExpr:
		result = append(result, expandExpr(typed, idents))
	}
	return result
}

func expandExpr(expr ast.Expr, idents map[*ast.Ident]types.Object) opchain {
	var result opchain
	for expr != nil {
		switch typed := expr.(type) {
		case *ast.CallExpr:
			var isNew bool
			var funcNode ast.Expr
			switch fTyped := typed.Fun.(type) {
			case *ast.FuncLit:
				expr = nil
				funcNode = fTyped
			case *ast.Ident:
				isNew = isNewObjectFunc(fTyped, idents)
				funcNode = fTyped
				expr = nil
			case *ast.SelectorExpr:
				isNew = isNewObjectFunc(fTyped.Sel, idents)
				funcNode = fTyped.Sel
				expr = fTyped.X
			}
			var op operation
			if isNew {
				op = &newOp{typ: typed.Args[0]}
			} else {
				eop := &execOp{fun: funcNode}
				for _, arg := range typed.Args {
					eop.args = append(eop.args, expandExpr(arg, idents))
				}
				op = eop
			}
			result = append(result, op)
		case *ast.SelectorExpr:
			result = append([]operation{newROpIdent(typed.Sel)}, result...)
			expr = typed.X
		case *ast.Ident:
			result = append([]operation{newROpIdent(typed)}, result...)
			expr = nil
		case *ast.IndexExpr:
			indexExpanded := expandExpr(typed.Index, idents)
			xExpanded := expandExpr(typed.X, idents)
			result = append(result, &indexOp{
				x:     xExpanded,
				index: indexExpanded,
			})
			expr = nil
		case *ast.BasicLit:
			result = append([]operation{newROpBasicLit(typed)}, result...)
			expr = nil
		case *ast.StarExpr:
			xExpanded := expandExpr(typed.X, idents)
			result = append(result, &derefOp{
				x: xExpanded,
			})
			expr = nil
		case *ast.CompositeLit:
			var inits []opchain
			for _, elt := range typed.Elts {
				inits = append(inits, expandExpr(elt, idents))
			}
			result = append(result, &newOp{
				typ:   typed.Type,
				inits: inits,
			})
			expr = nil
		default:
			fmt.Printf("expandExpr: skipping %s\n", reflect.ValueOf(expr).Type().String())
			expr = nil
		}
	}
	return result
}

func isNewObjectFunc(id *ast.Ident, idents map[*ast.Ident]types.Object) bool {
	if id.Obj == nil {
		return false
	}
	if id.Obj.Name != "new" && id.Obj.Name != "make" {
		return false
	}
	t, ok := idents[id]
	if !ok {
		return false
	}
	_, ok = t.(*types.Builtin)
	return ok
}

func exprName(expr ast.Expr) string {
	name := "<expr>"
	switch typed := expr.(type) {
	case *ast.Ident:
		name = typed.Name
	case *ast.SelectorExpr:
		name = typed.Sel.Name
	case *ast.BasicLit:
		name = typed.Value
	}
	return name
}

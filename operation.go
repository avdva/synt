// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"bytes"
	"fmt"
	"go/ast"
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

func statementsToOpchain(statements []ast.Stmt) opFlow {
	var result opFlow
	for _, statement := range statements {
		result = append(result, statementToOpchain(statement)...)
	}
	return result
}

func statementToOpchain(statement ast.Stmt) opFlow {
	var result opFlow
	switch typed := statement.(type) {
	case *ast.AssignStmt:
		result = append(result, expandAssignment(typed.Lhs, typed.Rhs)...)
	case *ast.ExprStmt:
		result = append(result, expandExprStmt(typed)...)
	case *ast.IncDecStmt:
		result = append(result, expandIncDec(typed)...)
	default:
		fmt.Printf("statementToOpchain: skipping %s\n", reflect.ValueOf(statement).Type().String())
	}
	return result
}

func expandIncDec(stmt *ast.IncDecStmt) opFlow {
	var result opFlow
	expanded := expandExpr(stmt.X)
	if len(expanded) > 1 {
		result = append(result, expanded[:len(expanded)-1])
	}
	return result
}

func expandAssignment(lhs, rhs []ast.Expr) opFlow {
	var result opFlow
	var writes []wOp
	for _, lhs := range lhs {
		expanded := expandExpr(lhs)
		if fl := flatten(expanded, 1); len(fl) > 1 {
			result = append(result, fl[:len(fl)-1])
		}
		writes = append(writes, wOp{lhs: expanded})
	}
	for i, rhs := range rhs {
		expanded := expandExpr(rhs)
		result = append(result, expanded)
		writes[i].rhs = expanded
	}
	for i := range writes {
		result = append(result, []operation{&writes[i]})
	}
	return result
}

func expandExprStmt(expr *ast.ExprStmt) opFlow {
	var result opFlow
	switch typed := expr.X.(type) {
	case *ast.CallExpr:
		result = append(result, expandExpr(typed))
	}
	return result
}

func expandExpr(expr ast.Expr) opchain {
	var result opchain
	for expr != nil {
		switch typed := expr.(type) {
		case *ast.CallExpr:
			var isNew bool
			switch fTyped := typed.Fun.(type) {
			case *ast.FuncLit:
				expr = nil
			case *ast.Ident:
				isNew = isNewObjectFunc(fTyped, nil)
				expr = nil
			case *ast.SelectorExpr:
				isNew = isNewObjectFunc(fTyped.Sel, nil)
				expr = fTyped.X
			}
			var op operation
			if isNew {
				op = &newOp{typ: typed.Args[0]}
			} else {
				eop := &execOp{fun: typed.Fun}
				for _, arg := range typed.Args {
					eop.args = append(eop.args, expandExpr(arg))
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
			indexExpanded := expandExpr(typed.Index)
			xExpanded := expandExpr(typed.X)
			result = append(result, &indexOp{
				x:     xExpanded,
				index: indexExpanded,
			})
			expr = nil
		case *ast.BasicLit:
			result = append([]operation{newROpBasicLit(typed)}, result...)
			expr = nil
		case *ast.StarExpr:
			xExpanded := expandExpr(typed.X)
			result = append(result, &derefOp{
				x: xExpanded,
			})
			expr = nil
		case *ast.CompositeLit:
			var inits []opchain
			for _, elt := range typed.Elts {
				inits = append(inits, expandExpr(elt))
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

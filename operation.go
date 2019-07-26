// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"bytes"
	"fmt"
	"go/ast"
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
		switch typed := statement.(type) {
		case *ast.AssignStmt:
			result = append(result, expandAssignment(typed.Lhs, typed.Rhs)...)
		case *ast.ExprStmt:
			result = append(result, expandExprStmt(typed))
		case *ast.IncDecStmt:
			result = append(result, expandIncDec(typed)...)
		default:
			fmt.Printf("statementsToOpchain: skipping %s\n", reflect.ValueOf(statement).Type().String())
		}
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
		if len(expanded) > 1 {
			result = append(result, expanded[:len(expanded)-1])
		}
		writes = append(writes, wOp{lhs: cleanAccessExpr(expanded)})
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

func expandExprStmt(expr *ast.ExprStmt) opchain {
	var result opchain
	switch typed := expr.X.(type) {
	case *ast.CallExpr:
		result = append(result, expandExpr(typed)...)
	}
	return result
}

func expandExpr(expr ast.Expr) opchain {
	var result opchain
	for expr != nil {
		switch typed := expr.(type) {
		case *ast.CallExpr:
			op := &execOp{fun: typed.Fun}
			for _, arg := range typed.Args {
				expanded := expandExpr(arg)
				result = append(result, expanded...)
				op.args = append(op.args, expanded)
			}
			switch fTyped := typed.Fun.(type) {
			case *ast.FuncLit:
				expr = nil
			case *ast.Ident:
				expr = nil
			case *ast.SelectorExpr:
				expr = fTyped.X
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
			result = append(result, indexExpanded...)
			result = append(result, xExpanded...)
			result = append(result, &execOp{
				args: []opchain{xExpanded, indexExpanded},
				fun:  ast.NewIdent("__indexaccess"),
			})
			expr = nil
		case *ast.BasicLit:
			result = append([]operation{newROpBasicLit(typed)}, result...)
			expr = nil
		case *ast.StarExpr:
			xExpanded := expandExpr(typed.X)
			result = append(result, xExpanded...)
			result = append(result, &execOp{
				args: []opchain{xExpanded},
				fun:  ast.NewIdent("__ptraccess"),
			})
			expr = nil
		default:
			fmt.Printf("expandExpr: skipping %s\n", reflect.ValueOf(expr).Type().String())
			expr = nil
		}
	}
	return result
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

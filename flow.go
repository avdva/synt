// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
)

type flowNode struct {
	statements []ast.Stmt
	branches   []flow
	gos        []ast.Stmt
	defers     []ast.Stmt
}

type flow struct {
	nodes []flowNode
}

func (fl *flow) nodeToAppend() *flowNode {
	needNew := len(fl.nodes) == 0
	if !needNew {
		last := fl.nodes[len(fl.nodes)-1]
		needNew = len(last.branches) > 0 || len(last.defers) > 0 || len(last.gos) > 0
	}
	if needNew {
		fl.nodes = append(fl.nodes, flowNode{})
	}
	return &fl.nodes[len(fl.nodes)-1]
}

func (fl *flow) addStatements(statements []ast.Stmt) {
	fn := fl.nodeToAppend()
	fn.statements = append(fn.statements, statements...)
}

func (fl *flow) addBranches(branches []flow) {
	fn := fl.nodeToAppend()
	fn.branches = branches
}

func (fl *flow) addDefers(defers []ast.Stmt) {
	fn := fl.nodeToAppend()
	fn.defers = append(fn.defers, defers...)
}

func (fl *flow) addGos(gos []ast.Stmt) {
	fn := fl.nodeToAppend()
	fn.defers = append(fn.gos, gos...)
}

func (fl *flow) addFlow(toAdd flow) {
	fl.nodes = append(fl.nodes, toAdd.nodes...)
}

func buildFlow(stmt *ast.BlockStmt) flow {
	var result flow
	for _, st := range stmt.List {
		switch typed := st.(type) {
		case *ast.IfStmt:
			result.addBranches(buildIfFlowNode(typed))
		case *ast.SwitchStmt:
		case *ast.GoStmt:
			result.addStatements(expressionsToStatemants(typed.Call.Args...))
			result.addGos(expressionsToStatemants(typed.Call))
		case *ast.TypeSwitchStmt, *ast.SelectStmt, *ast.RangeStmt:
		case *ast.BlockStmt:
			result.addBranches([]flow{buildFlow(typed)})
		case *ast.DeferStmt:
			result.addStatements(expressionsToStatemants(typed.Call.Args...))
			result.addDefers(expressionsToStatemants(typed.Call))
		default:
			result.addStatements([]ast.Stmt{st})
		}
	}
	return result
}

func buildIfFlowNode(stmt *ast.IfStmt) []flow {
	var inits []ast.Stmt
	var branches []flow
	for stmt != nil {
		var branchFlow flow
		inits = append(inits, stmt.Init)
		branchFlow.addStatements(expressionsToStatemants(stmt.Cond))
		branchFlow.addFlow(buildFlow(stmt.Body))
		branches = append(branches, branchFlow)
		switch typed := stmt.Else.(type) {
		case nil:
			stmt = nil
		case *ast.IfStmt:
			stmt = typed
		case *ast.BlockStmt:
			branches = append(branches, buildFlow(typed))
			stmt = nil
		case *ast.ExprStmt:
			var fl flow
			fl.addStatements([]ast.Stmt{typed})
			branches = append(branches, fl)
			stmt = nil
		}
	}
	for i := range inits {
		var initFl flow
		initFl.addStatements(inits[i:])
		initFl.addFlow(branches[i])
		branches[i] = initFl
	}
	return branches
}

func expressionsToStatemants(exprs ...ast.Expr) []ast.Stmt {
	var stmts []ast.Stmt
	for _, expr := range exprs {
		stmts = append(stmts, &ast.ExprStmt{X: expr})
	}
	return stmts
}

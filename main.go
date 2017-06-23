// Copyright 2017 Aleksandr Demakin. All rights reserved.

package main

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/avdva/synt/linter"

	log "github.com/Sirupsen/logrus"
)

func main() {
	var exitCode int
	flag.Parse()
	toParse := flag.Args()
	if len(toParse) == 0 {
		toParse = []string{"."}
	}
	for _, name := range toParse {
		if fi, err := os.Stat(name); err == nil && fi.IsDir() {
			if !doDir(name) {
				exitCode = 1
			}
		} else {
			log.Warnf("not a directory: %s", name)
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

func doDir(name string) bool {
	ok := true
	notests := func(info os.FileInfo) bool {
		return info.IsDir() || !strings.HasSuffix(info.Name(), "_test.go")
	}
	fs := token.NewFileSet()
	if pkgs, err := parser.ParseDir(fs, name, notests, parser.ParseComments); err != nil {
		log.Errorf("%s parsing error: %v", name, err)
		ok = false
	} else {
		for _, pkg := range pkgs {
			ok = doPackage(fs, pkg) && ok
		}
	}
	return ok
}

func doPackage(fs *token.FileSet, pkg *ast.Package) bool {
	l := linter.New(fs, pkg)
	reports := l.Do()
	//sort.Sort(reports)
	for _, report := range reports {
		_ = report
		//errorf("%s: %s is unused", fs.Position(report.Pos), report.Name)
	}
	return len(reports) == 0
}

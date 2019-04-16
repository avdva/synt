// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

var (
	stdLockers = []string{"sync.Mutex", "sync.RWMutex"}
)

type Linter struct {
	fs       *token.FileSet
	pkgs     map[string]*ast.Package
	checkers []string
}

type Report struct {
	Err      string
	Location string
}

func New(path string, checkers []string) (*Linter, error) {
	pkgs, fs, err := parsePackage(path)
	if err != nil {
		return nil, err
	}
	return &Linter{fs: fs, pkgs: pkgs, checkers: checkers}, nil
}

func DoDir(name string, checkers []string) ([]Report, error) {
	l, err := New(name, checkers)
	if err != nil {
		return nil, err
	}
	return l.Do("")
}

func (l *Linter) Do(pkg string) (result []Report, firstErr error) {
	checkers := makeCheckers(l.checkers)
	if len(pkg) == 0 {
		for name, pkg := range l.pkgs {
			if reports, err := doPackage(pkg, l.fs, checkers...); err != nil {
				if firstErr == nil {
					firstErr = errors.Wrapf(err, "error checking package %q", name)
				}
			} else {
				result = append(result, reports...)
			}
		}
	} else if pkg, found := l.pkgs[pkg]; found {
		result, firstErr = doPackage(pkg, l.fs, checkers...)
	} else {
		firstErr = errors.New("no such package")
	}
	return result, firstErr
}

func doPackage(pkg *ast.Package, fs *token.FileSet, checkers ...Checker) ([]Report, error) {
	var reports []CheckReport
	info := &CheckInfo{
		Pkg: pkg,
		Fs:  fs,
	}
	for _, checker := range checkers {
		if reps, err := checker.DoPackage(info); err == nil {
			reports = append(reports, reps...)
		} else {
			return nil, err
		}
	}
	return checkReportsToReports(reports, fs), nil
}

func checkReportsToReports(reports []CheckReport, fs *token.FileSet) []Report {
	if reports == nil {
		return nil
	}
	sort.Slice(reports, func(i, j int) bool {
		lhs, rhs := fs.Position(reports[i].Pos), fs.Position(reports[j].Pos)
		if lhs.Filename != rhs.Filename {
			return lhs.Filename < rhs.Filename
		}
		return reports[i].Pos < reports[j].Pos
	})
	var result []Report
	for _, e := range reports {
		result = append(result, Report{
			Err:      e.Err.Error(),
			Location: fs.Position(e.Pos).String(),
		})
	}
	return result
}

func parsePackage(path string) (map[string]*ast.Package, *token.FileSet, error) {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, path, notests, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return pkgs, fs, nil
}

func notests(info os.FileInfo) bool {
	return info.IsDir() || !strings.HasSuffix(info.Name(), "_test.go")
}

func makeCheckers(names []string) []Checker {
	var result []Checker
	for _, name := range names {
		if ch := makeChecker(name); ch != nil {
			result = append(result, ch)
		}
	}
	return result
}

func makeChecker(name string) Checker {
	switch name {
	case "m":
		return newMutexChecker()
	case "mstate":
		return newLockerStateChecker(lscConfig{lockTypes: stdLockers})
	default:
		return nil
	}
}

// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLockerStateCheckerDoubleLock(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testLocation, testPkg0Path}, "/")
	pkgs, fs, err := parsePackage(path)
	r.NoError(err)
	ch := newLockerStateChecker(stdLockers, makeLSCFilter("doubleLock"))
	actual, err := ch.DoPackage(&CheckInfo{Pkg: pkgs[testPkg0Package], Fs: fs})
	r.NoError(err)
	file := path + "/main.go"
	expected := []Report{
		{Location: file + ":23:4", Err: "cannot \"lock\"  [already locked]"},
	}
	r.Equal(expected, checkReportsToReports(actual, fs))
}

func TestLockerStateCheckerDoubleUnLock(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testLocation, testPkg0Path}, "/")
	pkgs, fs, err := parsePackage(path)
	r.NoError(err)
	ch := newLockerStateChecker(stdLockers, makeLSCFilter("doubleUnlock"))
	actual, err := ch.DoPackage(&CheckInfo{Pkg: pkgs[testPkg0Package], Fs: fs})
	r.NoError(err)
	file := path + "/main.go"
	expected := []Report{
		{Location: file + ":29:4", Err: "cannot \"unlock\"  [not locked]"},
	}
	r.Equal(expected, checkReportsToReports(actual, fs))
}

func TestLockerStateCheckerUnlockedUnLock(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testLocation, testPkg0Path}, "/")
	pkgs, fs, err := parsePackage(path)
	r.NoError(err)
	ch := newLockerStateChecker(stdLockers, makeLSCFilter("unlockedUnlock"))
	actual, err := ch.DoPackage(&CheckInfo{Pkg: pkgs[testPkg0Package], Fs: fs})
	r.NoError(err)
	file := path + "/main.go"
	expected := []Report{
		{Location: file + ":33:4", Err: "cannot \"unlock\"  [not locked]"},
	}
	r.Equal(expected, checkReportsToReports(actual, fs))
}

func TestLockerStateCheckerIfLock(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testLocation, testPkg0Path}, "/")
	pkgs, fs, err := parsePackage(path)
	r.NoError(err)
	ch := newLockerStateChecker(stdLockers, makeLSCFilter("ifLock"))
	actual, err := ch.DoPackage(&CheckInfo{Pkg: pkgs[testPkg0Package], Fs: fs})
	r.NoError(err)
	file := path + "/main.go"
	expected := []Report{
		{Location: file + ":39:5", Err: "cannot \"rlock\"  [already locked]"},
	}
	r.Equal(expected, checkReportsToReports(actual, fs))
}

func TestLockerStateCheckerIfLock2(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testLocation, testPkg0Path}, "/")
	pkgs, fs, err := parsePackage(path)
	r.NoError(err)
	ch := newLockerStateChecker(stdLockers, makeLSCFilter("ifLock2"))
	actual, err := ch.DoPackage(&CheckInfo{Pkg: pkgs[testPkg0Package], Fs: fs})
	r.NoError(err)
	var expected []Report
	r.Equal(expected, checkReportsToReports(actual, fs))
}

func TestLockerStateCheckerIfLock3(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testLocation, testPkg0Path}, "/")
	pkgs, fs, err := parsePackage(path)
	r.NoError(err)
	ch := newLockerStateChecker(stdLockers, makeLSCFilter("ifLock3"))
	actual, err := ch.DoPackage(&CheckInfo{Pkg: pkgs[testPkg0Package], Fs: fs})
	r.NoError(err)
	file := path + "/main.go"
	expected := []Report{
		{Location: file + ":58:4", Err: "cannot \"unlock\"  [?rwlocked]"},
	}
	r.Equal(expected, checkReportsToReports(actual, fs))
}

func TestLockerStateCheckerIfLock4(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testLocation, testPkg0Path}, "/")
	pkgs, fs, err := parsePackage(path)
	r.NoError(err)
	ch := newLockerStateChecker(stdLockers, makeLSCFilter("ifLock4"))
	actual, err := ch.DoPackage(&CheckInfo{Pkg: pkgs[testPkg0Package], Fs: fs})
	r.NoError(err)
	var expected []Report
	r.Equal(expected, checkReportsToReports(actual, fs))
}

func makeLSCFilter(names ...string) func(string) bool {
	return func(name string) bool {
		for _, n := range names {
			if n == name {
				return true
			}
		}
		return false
	}
}

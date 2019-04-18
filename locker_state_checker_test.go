// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func testFuncLSC(r *require.Assertions, info *CheckInfo, fn string, expected []Report) {
	ch := newLockerStateChecker(lscConfig{lockTypes: stdLockers, filter: makeLSCFilter(fn)})
	actual, err := ch.DoPackage(info)
	r.NoError(err)
	r.Equal(expected, checkReportsToReports(actual, info.Fs))
}

func TestLockerStateCheckerDoubleLock(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testPkg0Path, "main.go"}, "/")
	expected := []Report{
		{Location: path + ":23:4", Err: "cannot \"lock\"  [already locked]"},
	}
	testFuncLSC(r, pkg0CheckInfo, "doubleLock", expected)
}

func TestLockerStateCheckerDoubleUnLock(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testPkg0Path, "main.go"}, "/")
	expected := []Report{
		{Location: path + ":29:4", Err: "cannot \"unlock\"  [not locked]"},
	}
	testFuncLSC(r, pkg0CheckInfo, "doubleUnlock", expected)
}

func TestLockerStateCheckerUnlockedUnLock(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testPkg0Path, "main.go"}, "/")
	expected := []Report{
		{Location: path + ":33:4", Err: "cannot \"unlock\"  [not locked]"},
	}
	testFuncLSC(r, pkg0CheckInfo, "unlockedUnlock", expected)
}

func TestLockerStateCheckerIfLock(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testPkg0Path, "main.go"}, "/")
	expected := []Report{
		{Location: path + ":39:5", Err: "cannot \"rlock\"  [already locked]"},
	}
	testFuncLSC(r, pkg0CheckInfo, "ifLock", expected)
}

func TestLockerStateCheckerIfLock2(t *testing.T) {
	r := require.New(t)
	var expected []Report
	testFuncLSC(r, pkg0CheckInfo, "ifLock2", expected)
}

func TestLockerStateCheckerIfLock3(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testPkg0Path, "main.go"}, "/")
	expected := []Report{
		{Location: path + ":58:4", Err: "cannot \"unlock\"  [?rwlocked]"},
	}
	testFuncLSC(r, pkg0CheckInfo, "ifLock3", expected)
}

func TestLockerStateCheckerIfLock4(t *testing.T) {
	r := require.New(t)
	var expected []Report
	testFuncLSC(r, pkg0CheckInfo, "ifLock4", expected)
}

func TestLockerStateCheckerDefferedUnlock(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testPkg0Path, "main.go"}, "/")
	expected := []Report{
		{Location: path + ":81:10", Err: "cannot \"unlock\"  [not locked]"},
	}
	testFuncLSC(r, pkg0CheckInfo, "defferedUnlock", expected)
}

func TestLockerStateCheckerDefferedDoubleUnlock(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testPkg0Path, "main.go"}, "/")
	expected := []Report{
		{Location: path + ":86:10", Err: "cannot \"unlock\"  [not locked]"},
	}
	testFuncLSC(r, pkg0CheckInfo, "defferedDoubleUnlock", expected)
}

func TestLockerStateCheckerDefferedIfUnlock(t *testing.T) {
	r := require.New(t)
	path := strings.Join([]string{testPkg0Path, "main.go"}, "/")
	expected := []Report{
		{Location: path + ":94:10", Err: "cannot \"lock\"  [already ?locked]"},
		{Location: path + ":96:11", Err: "cannot \"runlock\"  [locked]"},
	}
	testFuncLSC(r, pkg0CheckInfo, "defferedIfUnlock", expected)
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

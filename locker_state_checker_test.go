// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLockerStateChecker(t *testing.T) {
	var path = strings.Join([]string{testLocation, test0Pkg, "main.go"}, "/")
	r := require.New(t)
	l, err := New("./testdata/pkg0", []string{"mstate"})
	r.NoError(err)
	reps, err := l.Do("")
	r.NoError(err)
	expected := []Report{
		{Location: path + ":23:4", Err: "cannot \"lock\"  [already locked]"},
		{Location: path + ":29:4", Err: "cannot \"unlock\"  [not locked]"},
		{Location: path + ":33:4", Err: "cannot \"unlock\"  [not locked]"},
	}
	r.Equal(expected, reps)
}

func TestLockerStateChecker2(t *testing.T) {
	r := require.New(t)
	l, err := New("./testdata/pkg1", []string{"mstate"})
	r.NoError(err)
	l.Do("")
}

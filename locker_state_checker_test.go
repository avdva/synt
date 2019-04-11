// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLockerStateChecker(t *testing.T) {
	r := require.New(t)
	l, err := New("./testdata/pkg0", []string{"mstate"})
	r.NoError(err)
	l.Do("")
}

func TestLockerStateChecker2(t *testing.T) {
	r := require.New(t)
	l, err := New("./testdata/pkg1", []string{"mstate"})
	r.NoError(err)
	l.Do("")
}

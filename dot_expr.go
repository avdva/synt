// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import "strings"

type dotExpr struct {
	parts []string
}

func dotExprFromParts(parts ...string) dotExpr {
	return dotExpr{parts: parts}
}

func (i dotExpr) String() string {
	return strings.Join(i.parts, ".")
}

func (i dotExpr) len() int {
	return len(i.parts)
}

func (i dotExpr) part(idx int) string {
	return i.parts[idx]
}

func (i dotExpr) eq(other dotExpr) bool {
	if len(i.parts) != len(other.parts) {
		return false
	}
	for i, p := range i.parts {
		if p != other.parts[i] {
			return false
		}
	}
	return true
}

func (i *dotExpr) field() dotExpr {
	return dotExpr{parts: []string{i.parts[len(i.parts)-1]}}
}

func (i dotExpr) selector() dotExpr {
	return dotExpr{parts: i.parts[:len(i.parts)-1]}
}

func (i dotExpr) first() dotExpr {
	return dotExpr{parts: []string{i.parts[0]}}
}

func (i dotExpr) copy() dotExpr {
	return dotExprFromParts(i.parts...)
}

func (i dotExpr) set(idx int, part string) {
	i.parts[idx] = part
}

func (i *dotExpr) append(part string) {
	i.parts = append(i.parts, part)
}

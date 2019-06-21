// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"strings"
)

type guard struct {
	object  dotExpr
	inverse bool
}

type annotation string

func (a annotation) guards() []guard {
	var result []guard
	parts := strings.Split(string(a), " ")
	for _, part := range parts {
		if strings.HasSuffix(part, ".Lock") {
			var g guard
			part = strings.TrimSuffix(part, ".Lock")
			if len(part) > 0 && part[0] == '!' {
				g.inverse = true
				part = part[1:]
			}
			g.object = dotExprFromParts(strings.Split(part, ".")...)
			result = append(result, g)
		}
	}
	return result
}

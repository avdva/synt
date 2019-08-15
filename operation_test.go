// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

import (
	"fmt"
	"go/ast"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"
)

func TestOperation1(t *testing.T) {
	r := require.New(t)
	defs := buildDefs(opTestCheckInfo.Pkg.Files)
	fdef := defs.functions["f1"]
	r.NotNil(fdef)
	desc, err := makePkgDesc(opTestCheckInfo.Pkg, opTestCheckInfo.Fs)
	r.NoError(err)
	of := statementsToOpchain(fdef.node.Body.List, desc.info.Uses)
	expected := opFlow{
		opchain{
			newROpBasicLit(&ast.BasicLit{Value: "5"}),
		},
		opchain{
			&wOp{
				lhs: opchain{newROpIdent(ast.NewIdent("a"))},
				rhs: opchain{newROpBasicLit(&ast.BasicLit{Value: "5"})},
			},
		},
		opchain{
			newROpIdent(ast.NewIdent("i")),
			newROpIdent(ast.NewIdent("sl")),
		},
		opchain{
			newROpBasicLit(&ast.BasicLit{Value: "8"}),
		},
		opchain{
			&wOp{
				lhs: opchain{
					&indexOp{
						index: opchain{&rOp{ast.NewIdent("i")}},
						x:     opchain{&rOp{ast.NewIdent("sl")}},
					},
				},
				rhs: opchain{newROpBasicLit(&ast.BasicLit{Value: "8"})},
			},
		},
		opchain{
			newROpIdent(ast.NewIdent("i")),
			newROpIdent(ast.NewIdent("sl")),
		},
		opchain{
			newROpIdent(ast.NewIdent("a")),
		},
		opchain{
			&wOp{
				lhs: opchain{
					&indexOp{
						index: opchain{&rOp{ast.NewIdent("i")}},
						x:     opchain{&rOp{ast.NewIdent("sl")}},
					},
				},
				rhs: opchain{newROpIdent(ast.NewIdent("a"))},
			},
		},

		// sl[sl[i]] = getSlice2(struct{}{})[getSlice(struct{}{})[0]]
		opchain{
			newROpIdent(ast.NewIdent("i")),
			newROpIdent(ast.NewIdent("sl")),
			&indexOp{
				index: opchain{&rOp{ast.NewIdent("i")}},
				x:     opchain{&rOp{ast.NewIdent("sl")}},
			},
			newROpIdent(ast.NewIdent("sl")),
		},
		opchain{
			newROpBasicLit(&ast.BasicLit{Value: "0"}),
			&newOp{
				typ: &ast.StructType{},
			},
			&execOp{
				fun: ast.NewIdent("getSlice"),
				args: []opchain{
					opchain{&newOp{
						typ: &ast.StructType{},
					}},
				},
			},
			&indexOp{
				index: opchain{newROpBasicLit(&ast.BasicLit{Value: "0"})},
				x: opchain{&execOp{
					fun: ast.NewIdent("getSlice"),
					args: []opchain{
						opchain{&newOp{
							typ: &ast.StructType{},
						}},
					},
				}},
			},
			&newOp{
				typ: &ast.StructType{},
			},
			&execOp{
				fun: ast.NewIdent("getSlice2"),
				args: []opchain{
					opchain{&newOp{
						typ: &ast.StructType{},
					}},
				},
			},
			&indexOp{
				index: opchain{&indexOp{
					index: opchain{newROpBasicLit(&ast.BasicLit{Value: "0"})},
					x: opchain{&execOp{
						fun: ast.NewIdent("getSlice"),
						args: []opchain{
							opchain{&newOp{
								typ: &ast.StructType{},
							}},
						},
					}},
				}},
				x: opchain{&execOp{
					fun: ast.NewIdent("getSlice2"),
					args: []opchain{
						opchain{&newOp{
							typ: &ast.StructType{},
						}},
					},
				}},
			},
		},
		opchain{
			&wOp{
				lhs: opchain{
					newROpIdent(ast.NewIdent("i")),
					newROpIdent(ast.NewIdent("sl")),
					&indexOp{
						index: opchain{&rOp{ast.NewIdent("i")}},
						x:     opchain{&rOp{ast.NewIdent("sl")}},
					},
					newROpIdent(ast.NewIdent("sl")),
				},
				rhs: opchain{
					newROpBasicLit(&ast.BasicLit{Value: "0"}),
					&newOp{
						typ: &ast.StructType{},
					},
					&execOp{
						fun: ast.NewIdent("getSlice"),
						args: []opchain{
							opchain{&newOp{
								typ: &ast.StructType{},
							}},
						},
					},
					&indexOp{
						index: opchain{newROpBasicLit(&ast.BasicLit{Value: "0"})},
						x: opchain{&execOp{
							fun: ast.NewIdent("getSlice"),
							args: []opchain{
								opchain{&newOp{
									typ: &ast.StructType{},
								}},
							},
						}},
					},
					&newOp{
						typ: &ast.StructType{},
					},
					&execOp{
						fun: ast.NewIdent("getSlice2"),
						args: []opchain{
							opchain{&newOp{
								typ: &ast.StructType{},
							}},
						},
					},
					&indexOp{
						index: opchain{&indexOp{
							index: opchain{newROpBasicLit(&ast.BasicLit{Value: "0"})},
							x: opchain{&execOp{
								fun: ast.NewIdent("getSlice"),
								args: []opchain{
									opchain{&newOp{
										typ: &ast.StructType{},
									}},
								},
							}},
						}},
						x: opchain{&execOp{
							fun: ast.NewIdent("getSlice2"),
							args: []opchain{
								opchain{&newOp{
									typ: &ast.StructType{},
								}},
							},
						}},
					},
				},
			},
		},
		opchain{
			newROpIdent(ast.NewIdent("a")),
		},
	}
	fmt.Println(of[:12])
	r.NoError(compareOpFlows(expected[:11], of[:11]))
}

func compareOpFlows(expected, given opFlow) error {
	var i int
	for i = 0; i < intMin(len(expected), len(given)); i++ {
		expectedStr, gotStr := expected[i].String(), flatten(given[i], -1).String()
		if expectedStr != gotStr {
			return errors.Errorf("expected, given: %s != %s", expectedStr, gotStr)
		}
	}
	if i < len(expected) {
		return errors.Errorf("rhs contains more ops: %s", expected[i:].String())
	}
	if i < len(given) {
		return errors.Errorf("lhs contains more ops: %s", given[i:].String())
	}
	return nil
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

type indexOp struct {
	x     opchain
	index opchain
}

func (op indexOp) Type() opType {
	return opExec
}

func (op indexOp) ObjectName() string {
	return op.x.ObjectName() + "[" + op.index.ObjectName() + "]"
}

func (op indexOp) String() string {
	return op.Type().String() + ":" + op.ObjectName()
}

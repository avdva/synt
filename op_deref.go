// Copyright 2019 Aleksandr Demakin. All rights reserved.

package synt

type derefOp struct {
	x opchain
}

func (op derefOp) Type() opType {
	return opExec
}

func (op derefOp) ObjectName() string {
	return "*(" + op.x.ObjectName() + ")"
}

func (op derefOp) String() string {
	return op.Type().String() + ":" + op.ObjectName()
}

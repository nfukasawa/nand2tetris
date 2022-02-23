package compiler

import (
	"fmt"
	"io"
)

type VMCmd string

const (
	VMCmdADD VMCmd = "add"
	VMCmdSUB VMCmd = "sub"
	VMCmdNEG VMCmd = "neg"
	VMCmdEQ  VMCmd = "eq"
	VMCmdGT  VMCmd = "gt"
	VMCmdLT  VMCmd = "lt"
	VMCmdAND VMCmd = "and"
	VMCmdOR  VMCmd = "or"
	VMCmdNOT VMCmd = "not"
)

type VMSeg string

const (
	VMSegCONST   VMSeg = "constant"
	VMSegARG     VMSeg = "argument"
	VMSegLOCAL   VMSeg = "local"
	VMSegSTATIC  VMSeg = "static"
	VMSegTHIS    VMSeg = "this"
	VMSegTHAT    VMSeg = "that"
	VMSegPOINTER VMSeg = "pointer"
	VMSegTEMP    VMSeg = "temp"
)

type JackVM struct {
	w   io.Writer
	err error
}

func NewJackVM(w io.Writer) *JackVM {
	return &JackVM{w: w}
}

func (vm *JackVM) Err() error {
	return vm.err
}

func (vm *JackVM) WritePush(seg VMSeg, idx int64) {
	vm.write(vm.w, "push %s %d\n", seg, idx)
}

func (vm *JackVM) WritePop(seg VMSeg, idx int64) {
	vm.write(vm.w, "pop %s %d\n", seg, idx)
}

func (vm *JackVM) WriteArithmetic(cmd VMCmd) {
	vm.write(vm.w, "%s\n", cmd)
}

func (vm *JackVM) WriteLabel(label string) {
	vm.write(vm.w, "label %s\n", label)
}

func (vm *JackVM) WriteGoto(label string) {
	vm.write(vm.w, "goto %s\n", label)
}

func (vm *JackVM) WriteIfGoto(label string) {
	vm.write(vm.w, "if-goto %s\n", label)
}

func (vm *JackVM) WriteCall(name string, args int64) {
	vm.write(vm.w, "call %s %d\n", name, args)
}

func (vm *JackVM) WriteFunction(name string, locals int64) {
	vm.write(vm.w, "function %s %d\n", name, locals)
}

func (vm *JackVM) WriteReturn() {
	vm.write(vm.w, "return\n")
}

func (vm *JackVM) write(w io.Writer, format string, a ...interface{}) {
	if vm.err != nil {
		return
	}
	_, vm.err = fmt.Fprintf(w, format, a...)
}

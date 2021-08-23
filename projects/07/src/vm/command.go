package vm

import "fmt"

type Command struct {
	Type CommandType

	// arithmetic command args
	ArithmeticOp ArithmeticOperation

	// memory access command args
	MemorySegment MemorySegment
	MemoryIndex   uint64
}

func (c *Command) String() string {
	switch c.Type {
	case CmdArithmetic:
		return string(c.ArithmeticOp)
	case CmdPush, CmdPop:
		return fmt.Sprintf("%s %s %d", c.Type, c.MemorySegment, c.MemoryIndex)
	default:
		return ""
	}
}

type CommandType string

const (
	CmdNone       CommandType = ""
	CmdArithmetic CommandType = "arith"
	CmdPush       CommandType = "push"
	CmdPop        CommandType = "pop"
	CmdLabel      CommandType = "TODO"
	CmdGoto       CommandType = "TODO"
	CmdIf         CommandType = "TODO"
	CmdFunction   CommandType = "TODO"
	CmdReturn     CommandType = "TODO"
	CmdCall       CommandType = "TODO"
)

type ArithmeticOperation string

const (
	OpNone ArithmeticOperation = ""
	OpAdd  ArithmeticOperation = "add"
	OpSub  ArithmeticOperation = "sub"
	OpNeg  ArithmeticOperation = "neg"
	OpEq   ArithmeticOperation = "eq"
	OpGt   ArithmeticOperation = "gt"
	OpLt   ArithmeticOperation = "lt"
	OpAnd  ArithmeticOperation = "and"
	OpOr   ArithmeticOperation = "or"
	OpNot  ArithmeticOperation = "not"
)

type MemorySegment string

const (
	SegNone     MemorySegment = ""
	SegArgument MemorySegment = "argument"
	SegLocal    MemorySegment = "local"
	SegStatic   MemorySegment = "static"
	SegConstant MemorySegment = "constant"
	SegThis     MemorySegment = "this"
	SegThat     MemorySegment = "that"
	SegPointer  MemorySegment = "pointer"
	SegTemp     MemorySegment = "temp"
)

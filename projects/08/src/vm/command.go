package vm

import (
	"fmt"
)

type Command struct {
	Type       CommandType
	Arithmetic *ArithmeticArgs
	Memory     *MemoryArgs
	Label      *LabelArgs
	Function   *FunctionArgs
}

type ArithmeticArgs struct {
	Operation ArithmeticOperation
}

type MemoryArgs struct {
	Segment MemorySegment
	Index   uint64
	Label   string
}

type LabelArgs struct {
	Label string
}

type FunctionArgs struct {
	Name string
	Num  uint64
}

func (c *Command) String() string {
	switch c.Type {
	case CmdArithmetic:
		return string(c.Arithmetic.Operation)
	case CmdPush, CmdPop:
		return fmt.Sprintf("%s %s %d", c.Type, c.Memory.Segment, c.Memory.Index)
	case CmdLabel, CmdGoto, CmdIfGoto:
		return fmt.Sprintf("%s %s", c.Type, c.Label.Label)
	case CmdFunction, CmdCall:
		return fmt.Sprintf("%s %s %d", c.Type, c.Function.Name, c.Function.Num)
	case CmdReturn:
		return string(CmdReturn)
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
	CmdLabel      CommandType = "label"
	CmdGoto       CommandType = "goto"
	CmdIfGoto     CommandType = "if-goto"
	CmdFunction   CommandType = "function"
	CmdReturn     CommandType = "return"
	CmdCall       CommandType = "call"
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

func CommandTypeFromString(str string) CommandType {
	switch str {
	case string(OpAdd), string(OpSub), string(OpNeg), string(OpEq), string(OpGt), string(OpLt), string(OpAnd), string(OpOr), string(OpNot):
		return CmdArithmetic
	case string(CmdPush), string(CmdPop), string(CmdLabel), string(CmdGoto), string(CmdIfGoto), string(CmdFunction), string(CmdReturn), string(CmdCall):
		return CommandType(str)
	default:
		return CmdNone
	}
}

func MemorySegmentFromString(str string) MemorySegment {
	switch str {
	case string(SegArgument), string(SegLocal), string(SegStatic), string(SegConstant), string(SegThis), string(SegThat), string(SegPointer), string(SegTemp):
		return MemorySegment(str)
	default:
		return SegNone
	}
}

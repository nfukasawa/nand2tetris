package vm

type CommandType int

const (
	CmdNone CommandType = iota
	CmdArithmetic
	CmdPush
	CmdPop
	CmdLabel
	CmdGoto
	CmdIf
	CmdFunction
	CmdReturn
	CmdCall
)

type ArithmeticOperation int

const (
	OpNone ArithmeticOperation = iota
	OpAdd
	OpSub
	OpNeg
	OpEq
	OpGt
	OpLt
	OpAnd
	OpOr
	OpNot
)

type MemorySegment int

const (
	SegNone MemorySegment = iota
	SegArgument
	SegLocal
	SegStatic
	SegConstant
	SegThis
	SegThat
	SegPointer
	SegTemp
)

type Command struct {
	Type CommandType

	// arithmetic command args
	ArithmeticOp ArithmeticOperation

	// memory access command args
	MemorySegment MemorySegment
	MemoryIndex   uint64
}

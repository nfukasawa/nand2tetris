package vm

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Translator struct {
	out   io.Writer
	Debug bool
}

func NewTranslator(out io.Writer) (*Translator, error) {
	return &Translator{out: out}, nil
}

func (t *Translator) File(fileName string) FileTranslator {
	return FileTranslator{
		fileName: fileName,
		opIndex:  0,
		out:      t.out,
		debug:    t.Debug,
	}
}

type FileTranslator struct {
	fileName string
	opIndex  int
	out      io.Writer
	debug    bool
}

func (t *FileTranslator) Command(cmd Command) error {
	if t.debug {
		t.writeAsm("// " + cmd.String())
	}
	switch cmd.Type {
	case CmdArithmetic:
		return t.arithmetic(cmd.ArithmeticOp)
	case CmdPush:
		return t.push(cmd.MemorySegment, cmd.MemoryIndex)
	case CmdPop:
		return t.pop(cmd.MemorySegment, cmd.MemoryIndex)
	default:
		return fmt.Errorf("unimplemented command type: %v", cmd.Type)
	}
}

func (t *FileTranslator) arithmetic(op ArithmeticOperation) error {
	switch op {
	case OpNeg, OpNot:
		return t.arith1Arg(op)
	case OpAdd, OpSub, OpAnd, OpOr:
		return t.arith2Args(op)
	case OpEq, OpGt, OpLt:
		return t.arithCmp(op)
	default:
		return fmt.Errorf("unknown arithmetic operation: %v", op)
	}
}

func (t *FileTranslator) arith1Arg(op ArithmeticOperation) error {
	var cmd string
	switch op {
	case OpNeg:
		cmd = "M=-M"
	case OpNot:
		cmd = "M=!M"
	}
	return t.writeAsm("@SP", "A=M-1", cmd) // y=sp-1; cmd(*y)
}

func (t *FileTranslator) arith2Args(op ArithmeticOperation) error {
	var cmd string
	switch op {
	case OpAdd:
		cmd = "M=M+D"
	case OpSub:
		cmd = "M=M-D"
	case OpAnd:
		cmd = "M=M&D"
	case OpOr:
		cmd = "M=M|D"
	}
	return t.writeAsm(
		"@SP", "M=M-1", "A=M", "D=M", "A=A-1", cmd, // y=--sp; x=y-1; cmd(*x, *y)
	)
}

func (t *FileTranslator) arithCmp(op ArithmeticOperation) error {
	label := fmt.Sprintf("CMP.%s.%d", t.fileName, t.opIndex)
	t.opIndex++

	var cmd string
	switch op {
	case OpEq:
		cmd = "D;JEQ"
	case OpGt:
		cmd = "D;JGT"
	case OpLt:
		cmd = "D;JLT"
	}
	return t.writeAsm(
		"@SP", "M=M-1", "A=M", "D=M", "A=A-1", // y=--sp; x=y-1
		"D=M-D",        // d=*x-*y
		"M=-1",         // *y=true
		"@"+label, cmd, // if cmd(d) goto label
		"@SP", "A=M-1", "M=0", // *y=false
		"("+label+")", // label
	)
}

func (t *FileTranslator) push(seg MemorySegment, index uint64) error {
	var cmds []string
	switch seg {
	case SegArgument, SegLocal, SegThis, SegThat:
		cmds = []string{t.reservedSegPos(seg), "D=M", "@" + toStr(index), "A=D+A", "D=M"}
	case SegPointer:
		cmds = []string{t.pointerSegPos(index), "D=M"}
	case SegTemp:
		cmds = []string{t.tempSegPos(index), "D=M"}
	case SegStatic:
		cmds = []string{t.staticSegPos(index), "D=M"}
	case SegConstant:
		cmds = []string{"@" + toStr(index), "D=A"}
	default:
		return fmt.Errorf("unknown push memory segment: %v", seg)
	}
	return t.writeAsm(append(cmds, "@SP", "M=M+1", "A=M-1", "M=D")...) // *sp=*pos; sp++;
}

func (t *FileTranslator) pop(seg MemorySegment, index uint64) error {
	switch seg {
	case SegArgument, SegLocal, SegThis, SegThat:
		return t.writeAsm(
			t.reservedSegPos(seg), "D=M", "@"+toStr(index), "D=D+A", "@R13", "M=D", // r13=pos+i
			"@SP", "M=M-1", "A=M", "D=M", // sp--; d=*sp
			"@R13", "A=M", "M=D", // *r13=d
		)
	}

	var cmd string
	switch seg {
	case SegPointer:
		cmd = t.pointerSegPos(index)
	case SegTemp:
		cmd = t.tempSegPos(index)
	case SegStatic:
		cmd = t.staticSegPos(index)
	default:
		return fmt.Errorf("unknown pop memory segment: %v", seg)
	}
	return t.writeAsm("@SP", "M=M-1", "A=M", "D=M", cmd, "M=D") // sp--; *pos=*sp
}

func (t *FileTranslator) reservedSegPos(seg MemorySegment) string {
	switch seg {
	case SegArgument:
		return "@ARG"
	case SegLocal:
		return "@LCL"
	case SegThis:
		return "@THIS"
	case SegThat:
		return "@THAT"
	default:
		return ""
	}
}

func (t *FileTranslator) pointerSegPos(index uint64) string {
	switch index {
	case 0:
		return "@THIS"
	case 1:
		return "@THAT"
	default:
		return ""
	}
}

func (t *FileTranslator) tempSegPos(index uint64) string {
	return "@R" + toStr(index+5)
}

func (t *FileTranslator) staticSegPos(index uint64) string {
	return "@" + t.fileName + "." + toStr(index)
}

func (t *FileTranslator) writeAsm(ops ...string) error {
	_, err := io.Copy(t.out, bytes.NewBufferString(strings.Join(ops, "\n")+"\n"))
	return err
}

func toStr(i uint64) string {
	return strconv.FormatUint(i, 10)
}

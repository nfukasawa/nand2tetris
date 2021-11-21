package vm

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Translator struct {
	out     io.Writer
	fileNum int
	Debug   bool
}

func NewTranslator(out io.Writer) (*Translator, error) {
	return &Translator{out: out, fileNum: 0}, nil
}

func (t *Translator) File(fileName string) FileTranslator {
	defer func() { t.fileNum++ }()

	return FileTranslator{
		fileName:   fileName,
		labelIndex: 0,
		out:        t.out,
		debug:      t.Debug,
	}
}

type FileTranslator struct {
	fileName     string
	functionName string
	labelIndex   int
	out          io.Writer
	debug        bool
}

func (t *FileTranslator) Command(cmd Command) error {
	if t.debug {
		t.writeAsm("// " + cmd.String())
	}
	switch cmd.Type {
	case CmdArithmetic:
		return t.arithmetic(cmd.Arithmetic)
	case CmdPush:
		return t.push(cmd.Memory)
	case CmdPop:
		return t.pop(cmd.Memory)
	case CmdLabel:
		return t.label(cmd.Label)
	case CmdGoto:
		return t.goTo(cmd.Label)
	case CmdIfGoto:
		return t.ifGoTo(cmd.Label)
	case CmdFunction:
		return t.function(cmd.Function)
	case CmdReturn:
		return t.ret()
	case CmdCall:
		return t.call(cmd.Function)
	default:
		return fmt.Errorf("unimplemented command type: %v", cmd.Type)
	}
}

func (t *FileTranslator) arithmetic(args *ArithmeticArgs) error {
	switch op := args.Operation; op {
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
	label := t.uniqueLabel("CMP")

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
		"@"+label, cmd, // if cmd(d) goto @label
		"@SP", "A=M-1", "M=0", // *y=false
		"("+label+")", // label
	)
}

func (t *FileTranslator) push(args *MemoryArgs) error {
	seg := args.Segment
	index := args.Index

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

func (t *FileTranslator) pop(args *MemoryArgs) error {
	seg := args.Segment
	index := args.Index

	switch seg {
	case SegArgument, SegLocal, SegThis, SegThat:
		return t.writeAsm(
			t.reservedSegPos(seg), "D=M", "@"+toStr(index), "D=D+A", "@R13", "M=D", // r13=pos+i
			"@SP", "M=M-1", "A=M", "D=M", // sp--
			"@R13", "A=M", "M=D", // *r13=*sp
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

func (t *FileTranslator) label(args *LabelArgs) error {
	return t.writeAsm("(" + t.scopedLabel(args) + ")") // (label)
}

func (t *FileTranslator) goTo(args *LabelArgs) error {
	return t.writeAsm("@"+t.scopedLabel(args), "0;JMP") // goto @label
}

func (t *FileTranslator) ifGoTo(args *LabelArgs) error {
	return t.writeAsm(
		"@SP", "M=M-1", "A=M", "D=M", // sp--
		"@"+t.scopedLabel(args), "D;JNE", // if(*sp != 0) goto @label
	)
}

func (t *FileTranslator) function(args *FunctionArgs) error {
	t.functionName = args.Name

	asm := []string{"(" + args.Name + ")"}  // (f)
	for i := uint64(0); i < args.Num; i++ { // repeat num times
		asm = append(asm, "@SP", "M=M+1", "A=M-1", "M=0") // sp++; sp=0
	}
	return t.writeAsm(asm...)
}

func (t *FileTranslator) ret() error {
	t.functionName = ""

	// FRAME = LCL
	// RET = *(FRAME-5)
	// *ARG = pop()
	// SP = ARG+1
	// THAT = *(FRAME-1)
	// THIS = *(FRAME-2)
	// ARG = *(FRAME-3)
	// LCL = *(FRAME-4)
	// goto RET
	return nil
}

func (t *FileTranslator) call(args *FunctionArgs) error {
	// label := t.uniqueLabel("RET")

	// push return-address
	// push LCL
	// push ARG
	// push THIS
	// push THAT
	// ARG = SP-n-5
	// LCL = SP
	// goto f
	// (return-address)
	return nil
}

func (t *FileTranslator) uniqueLabel(namespace string) string {
	label := fmt.Sprintf("%s.%s.%d", namespace, t.fileName, t.labelIndex)
	t.labelIndex++
	return label
}

func (t *FileTranslator) scopedLabel(args *LabelArgs) string {
	return t.functionName + "$" + args.Label
}

func (t *FileTranslator) writeAsm(ops ...string) error {
	_, err := io.Copy(t.out, bytes.NewBufferString(strings.Join(ops, "\n")+"\n"))
	return err
}

func concat(elms []interface{}) []string {
	var strs []string
	for _, elm := range elms {
		switch elm := elm.(type) {
		case string:
			strs = append(strs, elm)
		case []string:
			strs = append(strs, elm...)
		default:
		}
	}
	return strs
}

func toStr(i uint64) string {
	return strconv.FormatUint(i, 10)
}

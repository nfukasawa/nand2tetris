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
	var asm []string
	switch cmd.Type {
	case CmdArithmetic:
		asm = t.arithmetic(cmd.Arithmetic)
	case CmdPush:
		asm = t.push(cmd.Memory)
	case CmdPop:
		asm = t.pop(cmd.Memory)
	case CmdLabel:
		asm = t.label(cmd.Label)
	case CmdGoto:
		asm = t.goTo(cmd.Label)
	case CmdIfGoto:
		asm = t.ifGoTo(cmd.Label)
	case CmdFunction:
		asm = t.function(cmd.Function)
	case CmdReturn:
		asm = t.ret()
	case CmdCall:
		asm = t.call(cmd.Function)
	default:
		return fmt.Errorf("unimplemented command type: %v", cmd.Type)
	}

	if t.debug {
		asm = append([]string{"// " + cmd.String()}, asm...)
	}

	return t.writeAsm(asm...)
}

func (t *FileTranslator) arithmetic(args *ArithmeticArgs) []string {
	switch op := args.Operation; op {
	case OpNeg, OpNot:
		return t.arith1Arg(op)
	case OpAdd, OpSub, OpAnd, OpOr:
		return t.arith2Args(op)
	case OpEq, OpGt, OpLt:
		return t.arithCmp(op)
	default:
		fmt.Printf("unknown arithmetic operation: %v", op)
		return nil
	}
}

func (t *FileTranslator) arith1Arg(op ArithmeticOperation) []string {
	var cmd string
	switch op {
	case OpNeg:
		cmd = "M=-M"
	case OpNot:
		cmd = "M=!M"
	}
	return []string{
		"@SP", "A=M-1", cmd, // y=sp-1; cmd(*y)
	}

}

func (t *FileTranslator) arith2Args(op ArithmeticOperation) []string {
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
	return []string{
		"@SP", "M=M-1", "A=M", "D=M", "A=A-1", cmd, // y=--sp; x=y-1; cmd(*x, *y)
	}
}

func (t *FileTranslator) arithCmp(op ArithmeticOperation) []string {
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
	return []string{
		"@SP", "M=M-1", "A=M", "D=M", "A=A-1", // y=--sp; x=y-1
		"D=M-D",          // d=*x-*y
		"M=-1",           // *y=true
		"@" + label, cmd, // if cmd(d) goto @label
		"@SP", "A=M-1", "M=0", // *y=false
		"(" + label + ")", // label
	}
}

func (t *FileTranslator) push(args *MemoryArgs) []string {
	seg := args.Segment
	index := args.Index

	var cmds []string
	switch seg {
	case SegArgument, SegLocal, SegThis, SegThat:
		cmds = []string{t.reservedSegPos(seg), "D=M", "@" + t.pushLabel(args), "A=D+A", "D=M"}
	case SegPointer:
		cmds = []string{t.pointerSegPos(index), "D=M"}
	case SegTemp:
		cmds = []string{t.tempSegPos(index), "D=M"}
	case SegStatic:
		cmds = []string{t.staticSegPos(index), "D=M"}
	case SegConstant:
		cmds = []string{"@" + t.pushLabel(args), "D=A"}
	default:
		fmt.Printf("unknown push memory segment: %v", seg)
		return nil
	}
	return append(
		cmds, "@SP", "M=M+1", "A=M-1", "M=D", // *sp=*pos; sp++;
	)
}

func (t *FileTranslator) pushLabel(args *MemoryArgs) string {
	if args.Label != "" {
		return args.Label
	}
	return toStr(args.Index)
}

func (t *FileTranslator) pop(args *MemoryArgs) []string {
	seg := args.Segment
	index := args.Index

	var pos string
	switch seg {
	case SegArgument, SegLocal, SegThis, SegThat:
		pos = t.reservedSegPos(seg)
		if index != 0 {
			return []string{
				pos, "D=M", "@" + toStr(index), "D=D+A", "@R13", "M=D", // r13=pos+i
				"@SP", "M=M-1", "A=M", "D=M", // sp--
				"@R13", "A=M", "M=D", // *r13=*sp
			}
		}
	case SegPointer:
		pos = t.pointerSegPos(index)
	case SegTemp:
		pos = t.tempSegPos(index)
	case SegStatic:
		pos = t.staticSegPos(index)
	default:
		fmt.Printf("unknown pop memory segment: %v", seg)
		return nil
	}
	return []string{
		"@SP", "M=M-1", "A=M", "D=M", pos, "M=D", // sp--; *pos=*sp
	}
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

func (t *FileTranslator) label(args *LabelArgs) []string {
	return []string{
		"(" + t.scopedLabel(args) + ")", // (label)
	}
}

func (t *FileTranslator) goTo(args *LabelArgs) []string {
	return []string{
		"@" + t.scopedLabel(args), "0;JMP", // goto @label
	}
}

func (t *FileTranslator) ifGoTo(args *LabelArgs) []string {
	return []string{
		"@SP", "M=M-1", "A=M", "D=M", // sp--
		"@" + t.scopedLabel(args), "D;JNE", // if(*sp != 0) goto @label
	}
}

func (t *FileTranslator) function(args *FunctionArgs) []string {
	t.functionName = args.Name

	asm := []string{"(" + args.Name + ")"}  // (f)
	for i := uint64(0); i < args.Num; i++ { // repeat num times
		asm = append(asm, "@SP", "M=M+1", "A=M-1", "M=0") // sp++; sp=0
	}
	return asm
}

func (t *FileTranslator) ret() []string {
	t.functionName = ""

	framePos := "@14"
	retPos := "@15"

	return concat(
		[]string{
			"@LCL", "D=M", framePos, "M=D", // FRAME = LCL
			"@5", "A=D-A", "D=M", retPos, "M=D", // RET = *(FRAME-5)
		},
		t.pop(&MemoryArgs{Segment: SegArgument, Index: 0}), // *ARG = pop()
		[]string{
			"@ARG", "D=M+1", "@SP", "M=D", // SP = ARG+1
		},
		t.popFrame(framePos, SegThat),     // THAT = *(FRAME-1)
		t.popFrame(framePos, SegThis),     // THIS = *(FRAME-2)
		t.popFrame(framePos, SegArgument), // ARG = *(FRAME-3)
		t.popFrame(framePos, SegLocal),    // LCL = *(FRAME-4)
		[]string{
			retPos, "A=M", "0;JMP", // goto RET
		},
	)
}

func (t *FileTranslator) popFrame(framePos string, seg MemorySegment) []string {
	return []string{
		framePos,
		"D=M-1",
		"AM=D",
		"D=M",
		t.reservedSegPos(seg),
		"M=D",
	}
}

func (t *FileTranslator) call(args *FunctionArgs) []string {
	label := t.uniqueLabel("RET")

	return concat(
		t.push(&MemoryArgs{Segment: SegConstant, Label: label}), // push return-address
		t.push(&MemoryArgs{Segment: SegLocal, Label: "SP"}),     // push LCL
		t.push(&MemoryArgs{Segment: SegArgument, Label: "SP"}),  // push ARG
		t.push(&MemoryArgs{Segment: SegThis, Label: "SP"}),      // push THIS
		t.push(&MemoryArgs{Segment: SegThat, Label: "SP"}),      // push THAT
		[]string{
			"@SP", "D=M", "@" + toStr(5+args.Num), "D=D-A", t.reservedSegPos(SegArgument), "M=D", // ARG = SP-n-5
			"@SP", "D=M", "@LCL", "M=D", // LCL = SP
			"@" + args.Name, "0;JMP", // goto f
			"(" + label + ")", // (return-address)
		},
	)
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

func concat(ss ...[]string) []string {
	var ret []string
	for _, s := range ss {
		ret = append(ret, s...)
	}
	return ret
}

func toStr(i uint64) string {
	return strconv.FormatUint(i, 10)
}

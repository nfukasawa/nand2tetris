package vm

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Translator struct {
	out        io.Writer
	labelIndex int64
	Debug      bool
}

func NewTranslator(out io.Writer, bootstrap bool) (*Translator, error) {
	t := &Translator{out: out}
	if bootstrap {
		if err := t.writeAsm(t.bootstrap()...); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (t *Translator) File(fileName string) *FileTranslator {
	return &FileTranslator{
		fileName:   fileName,
		translator: t,
	}
}

func (t *Translator) command(cmd Command, file *FileTranslator) error {

	var asm []string
	switch cmd.Type {
	case CmdArithmetic:
		asm = t.arithmetic(cmd.Arithmetic, file)
	case CmdPush:
		asm = t.push(cmd.Memory, file)
	case CmdPop:
		asm = t.pop(cmd.Memory, file)
	case CmdLabel:
		asm = t.label(cmd.Label, file)
	case CmdGoto:
		asm = t.goTo(cmd.Label, file)
	case CmdIfGoto:
		asm = t.ifGoTo(cmd.Label, file)
	case CmdFunction:
		asm = t.function(cmd.Function, file)
	case CmdReturn:
		asm = t.ret(file)
	case CmdCall:
		asm = t.call(cmd.Function, file)
	default:
		return fmt.Errorf("unimplemented command type: %v", cmd.Type)
	}

	if t.Debug {
		asm = append([]string{"// " + cmd.String()}, asm...)
	}

	return t.writeAsm(asm...)
}

func (t *Translator) writeAsm(ops ...string) error {
	_, err := io.Copy(t.out, bytes.NewBufferString(strings.Join(ops, "\n")+"\n"))
	return err
}

func (t *Translator) bootstrap() []string {
	return concat(
		[]string{"@256", "D=A", "@SP", "M=D"},        // SP=256
		t.call(&FunctionArgs{Name: "Sys.init"}, nil), // call Sys.init
	)
}

func (t *Translator) arithmetic(args *ArithmeticArgs, file *FileTranslator) []string {
	switch op := args.Operation; op {
	case OpNeg, OpNot:
		return t.arith1Arg(op)
	case OpAdd, OpSub, OpAnd, OpOr:
		return t.arith2Args(op)
	case OpEq, OpGt, OpLt:
		return t.arithCmp(op, file)
	default:
		fmt.Printf("unknown arithmetic operation: %v", op)
		return nil
	}
}

func (t *Translator) arith1Arg(op ArithmeticOperation) []string {
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

func (t *Translator) arith2Args(op ArithmeticOperation) []string {
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

func (t *Translator) arithCmp(op ArithmeticOperation, file *FileTranslator) []string {
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

func (t *Translator) push(args *MemoryArgs, file *FileTranslator) []string {
	seg := args.Segment
	index := args.Index

	var cmds []string
	switch seg {
	case SegArgument, SegLocal, SegThis, SegThat:
		label := t.pushLabel(args)
		if label != "SP" {
			cmds = []string{t.reservedSegPos(seg), "D=M", "@" + label, "A=D+A", "D=M"}
		} else {
			cmds = []string{t.reservedSegPos(seg), "D=M"}
		}
	case SegPointer:
		cmds = []string{t.pointerSegPos(index), "D=M"}
	case SegTemp:
		cmds = []string{t.tempSegPos(index), "D=M"}
	case SegStatic:
		cmds = []string{t.staticSegPos(index, file), "D=M"}
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

func (t *Translator) pushLabel(args *MemoryArgs) string {
	if args.Label != "" {
		return args.Label
	}
	return toStr(args.Index)
}

func (t *Translator) pop(args *MemoryArgs, file *FileTranslator) []string {
	seg := args.Segment
	index := args.Index

	var pos string
	switch seg {
	case SegArgument, SegLocal, SegThis, SegThat:
		return []string{
			t.reservedSegPos(seg), "D=M", "@" + toStr(index), "D=D+A", "@R13", "M=D", // r13=pos+i
			"@SP", "M=M-1", "A=M", "D=M", // sp--
			"@R13", "A=M", "M=D", // *r13=*sp
		}
	case SegPointer:
		pos = t.pointerSegPos(index)
	case SegTemp:
		pos = t.tempSegPos(index)
	case SegStatic:
		pos = t.staticSegPos(index, file)
	default:
		fmt.Printf("unknown pop memory segment: %v", seg)
		return nil
	}
	return []string{
		"@SP", "M=M-1", "A=M", "D=M", pos, "M=D", // sp--; *pos=*sp
	}
}

func (t *Translator) reservedSegPos(seg MemorySegment) string {
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

func (t *Translator) pointerSegPos(index uint64) string {
	switch index {
	case 0:
		return "@THIS"
	case 1:
		return "@THAT"
	default:
		return ""
	}
}

func (t *Translator) tempSegPos(index uint64) string {
	return "@R" + toStr(index+5)
}

func (t *Translator) staticSegPos(index uint64, file *FileTranslator) string {
	return "@" + file.fileName + "." + toStr(index)
}

func (t *Translator) label(args *LabelArgs, file *FileTranslator) []string {
	return []string{
		"(" + file.scopedLabel(args) + ")", // (label)
	}
}

func (t *Translator) goTo(args *LabelArgs, file *FileTranslator) []string {
	return []string{
		"@" + file.scopedLabel(args), "0;JMP", // goto @label
	}
}

func (t *Translator) ifGoTo(args *LabelArgs, file *FileTranslator) []string {
	return []string{
		"@SP", "M=M-1", "A=M", "D=M", // sp--
		"@" + file.scopedLabel(args), "D;JNE", // if(*sp != 0) goto @label
	}
}

func (t *Translator) function(args *FunctionArgs, file *FileTranslator) []string {
	file.functionName = args.Name

	asm := []string{"(" + args.Name + ")"} // (f)
	for i := uint64(0); i < args.Num; i++ {
		asm = append(asm, "@SP", "M=M+1", "A=M-1", "M=0") // *sp=0; sp++;
	}
	return asm
}

func (t *Translator) ret(file *FileTranslator) []string {
	framePos := "@14"
	retPos := "@15"

	return concat(
		[]string{
			"@LCL", "D=M", framePos, "M=D", // FRAME = LCL
			"@5", "A=D-A", "D=M", retPos, "M=D", // RET = *(FRAME-5)
			"@SP", "M=M-1", "A=M", "D=M", t.reservedSegPos(SegArgument), "A=M", "M=D", // *ARG = pop()
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

func (t *Translator) popFrame(framePos string, seg MemorySegment) []string {
	return []string{
		framePos,
		"D=M-1",
		"AM=D",
		"D=M",
		t.reservedSegPos(seg),
		"M=D",
	}
}

func (t *Translator) call(args *FunctionArgs, file *FileTranslator) []string {
	retAddr := t.uniqueLabel("RET")

	return concat(
		t.push(&MemoryArgs{Segment: SegConstant, Label: retAddr}, file), // push return-address
		t.push(&MemoryArgs{Segment: SegLocal, Label: "SP"}, file),       // push LCL
		t.push(&MemoryArgs{Segment: SegArgument, Label: "SP"}, file),    // push ARG
		t.push(&MemoryArgs{Segment: SegThis, Label: "SP"}, file),        // push THIS
		t.push(&MemoryArgs{Segment: SegThat, Label: "SP"}, file),        // push THAT
		[]string{
			"@SP", "D=M", "@" + toStr(5+args.Num), "D=D-A", t.reservedSegPos(SegArgument), "M=D", // ARG = SP-n-5
			"@SP", "D=M", "@LCL", "M=D", // LCL = SP
			"@" + args.Name, "0;JMP", // goto f
			"(" + retAddr + ")", // (return-address)
		},
	)
}

func (t *Translator) uniqueLabel(namespace string) string {
	label := fmt.Sprintf("%s.%d", namespace, t.labelIndex)
	t.labelIndex++
	return label
}

type FileTranslator struct {
	fileName     string
	functionName string

	translator *Translator
}

func (t *FileTranslator) Command(cmd Command) error {
	if t == nil {
		return nil
	}
	return t.translator.command(cmd, t)
}

func (t *FileTranslator) scopedLabel(args *LabelArgs) string {
	if t == nil {
		return "$" + args.Label
	}

	return t.functionName + "$" + args.Label
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

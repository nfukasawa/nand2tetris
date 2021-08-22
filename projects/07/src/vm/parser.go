package vm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Parser struct {
	scanner *bufio.Scanner
	src     string
	line    int
}

func NewParser(src string) (*Parser, error) {
	file, err := os.Open(src)
	if err != nil {
		return nil, err
	}

	return &Parser{
		scanner: bufio.NewScanner(file),
		src:     src,
		line:    0,
	}, nil
}

func (p *Parser) NextCommand() (cmd Command, err error) {
	args, line, err := p.nextLine()
	if errors.Is(err, io.EOF) {
		return cmd, err
	}
	if err != nil {
		return cmd, fmt.Errorf("error %s:%d: %v", p.src, line, err)
	}
	cmd, err = p.mapCommand(args)
	if err != nil {
		return cmd, fmt.Errorf("error %s:%d: %v", p.src, line, err)
	}
	return cmd, err
}

func (p *Parser) nextLine() (args []string, line int, err error) {
	for {
		p.line++
		if !p.scanner.Scan() {
			return nil, p.line, io.EOF
		}

		if err := p.scanner.Err(); err != nil {
			return nil, p.line, err
		}

		cmd := strings.Trim(p.scanner.Text(), " ")
		if cmd == "" || strings.HasPrefix(cmd, "//") {
			continue
		}

		for _, arg := range strings.Split(cmd, " ") {
			if arg == "" {
				continue
			}
			if strings.HasPrefix(arg, "//") {
				break
			}
			args = append(args, arg)
		}
		break
	}

	return args, p.line, nil
}

func (p *Parser) mapCommand(args []string) (cmd Command, err error) {
	if len(args) == 0 {
		return cmd, fmt.Errorf("command empty")
	}
	switch args[0] {

	// arithmetic
	case "add":
		return p.mapArithmeticCommand(OpAdd, args[1:])
	case "sub":
		return p.mapArithmeticCommand(OpSub, args[1:])
	case "neg":
		return p.mapArithmeticCommand(OpNeg, args[1:])
	case "eq":
		return p.mapArithmeticCommand(OpEq, args[1:])
	case "gt":
		return p.mapArithmeticCommand(OpGt, args[1:])
	case "lt":
		return p.mapArithmeticCommand(OpLt, args[1:])
	case "and":
		return p.mapArithmeticCommand(OpAnd, args[1:])
	case "or":
		return p.mapArithmeticCommand(OpOr, args[1:])
	case "not":
		return p.mapArithmeticCommand(OpNot, args[1:])

	// memory access
	case "push":
		return p.mapMemoryCommand(CmdPush, args[1:])
	case "pop":
		return p.mapMemoryCommand(CmdPop, args[1:])

	default:
		return cmd, fmt.Errorf("unknown command")
	}
}

func (p *Parser) mapArithmeticCommand(op ArithmeticOperation, args []string) (cmd Command, err error) {
	if len(args) != 0 {
		return cmd, fmt.Errorf("arithmetic command takes no arguments")
	}
	return Command{Type: CmdArithmetic, ArithmeticOp: op}, nil
}

func (p *Parser) mapMemoryCommand(ty CommandType, args []string) (cmd Command, err error) {
	if len(args) != 2 {
		return cmd, fmt.Errorf("memory command takes 2 arguments")
	}

	var seg MemorySegment
	switch args[0] {
	case "argument":
		seg = SegArgument
	case "local":
		seg = SegLocal
	case "static":
		seg = SegStatic
	case "constant":
		seg = SegConstant
	case "this":
		seg = SegThis
	case "that":
		seg = SegThat
	case "pointer":
		seg = SegPointer
	case "temp":
		seg = SegTemp
	default:
		return cmd, fmt.Errorf("unknown memory segment: %s", args[0])
	}

	index, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return cmd, fmt.Errorf("invalid index: %s", args[1])
	}

	return Command{Type: ty, MemorySegment: seg, MemoryIndex: index}, nil
}

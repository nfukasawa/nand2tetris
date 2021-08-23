package vm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Parser struct {
	scanner *bufio.Scanner
	src     string
	line    int
}

func NewParser(srcPath string, r io.Reader) *Parser {
	return &Parser{
		scanner: bufio.NewScanner(r),
		src:     srcPath,
		line:    0,
	}
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
	case "add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not":
		return p.mapArithmeticCommand(ArithmeticOperation(args[0]), args[1:])

	// memory access
	case "push":
		return p.mapMemoryCommand(CmdPush, args[1:])
	case "pop":
		return p.mapMemoryCommand(CmdPop, args[1:])

	default:
		return cmd, fmt.Errorf("unknown command: %s", args[0])
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
	case "argument", "local", "static", "constant", "this", "that", "pointer", "temp":
		seg = MemorySegment(args[0])
	default:
		return cmd, fmt.Errorf("unknown memory segment: %s", args[0])
	}

	index, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return cmd, fmt.Errorf("invalid index: %s", args[1])
	}

	// validations
	if seg == SegPointer && index != 0 && index != 1 {
		return cmd, fmt.Errorf("pointer index must be 0 or 1")
	}
	if seg == SegTemp && index > 6 {
		return cmd, fmt.Errorf("temp index must be less than 7")
	}
	if ty == CmdPop && seg == SegConstant {
		return cmd, fmt.Errorf("pop command does not accept constant segment")
	}

	return Command{Type: ty, MemorySegment: seg, MemoryIndex: index}, nil
}

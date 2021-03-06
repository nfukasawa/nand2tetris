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

		line := p.scanner.Text()
		comment := strings.Index(line, "//")
		if comment != -1 {
			line = line[0:comment]
		}

		cmd := strings.TrimSpace(line)
		if cmd == "" {
			continue
		}
		for _, arg := range strings.Split(cmd, " ") {
			if arg == "" {
				continue
			}
			args = append(args, strings.TrimSpace(arg))
		}
		break
	}

	return args, p.line, nil
}

func (p *Parser) mapCommand(args []string) (cmd Command, err error) {
	if len(args) == 0 {
		return cmd, fmt.Errorf("command empty")
	}

	switch ty := CommandTypeFromString(args[0]); ty {
	case CmdArithmetic:
		return p.mapArithmeticCommand(ArithmeticOperation(args[0]), args[1:])
	case CmdPush, CmdPop:
		return p.mapMemoryCommand(ty, args[1:])
	case CmdLabel, CmdGoto, CmdIfGoto:
		return p.mapLabelCommand(ty, args[1:])
	case CmdFunction, CmdCall:
		return p.mapFunctionCommand(ty, args[1:])
	case CmdReturn:
		return Command{Type: ty}, nil
	default:
		return cmd, fmt.Errorf("unknown command: %s", args[0])
	}
}

func (p *Parser) mapArithmeticCommand(op ArithmeticOperation, args []string) (cmd Command, err error) {
	if len(args) != 0 {
		return cmd, fmt.Errorf("%s command takes no arguments", op)
	}
	return Command{
		Type: CmdArithmetic,
		Arithmetic: &ArithmeticArgs{
			Operation: op,
		},
	}, nil
}

func (p *Parser) mapMemoryCommand(ty CommandType, args []string) (cmd Command, err error) {
	if len(args) != 2 {
		return cmd, fmt.Errorf("%s command takes 2 arguments", ty)
	}

	seg := MemorySegmentFromString(args[0])
	if seg == SegNone {
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

	return Command{
		Type: ty,
		Memory: &MemoryArgs{
			Segment: seg,
			Index:   index,
		},
	}, nil
}

func (p *Parser) mapLabelCommand(ty CommandType, args []string) (cmd Command, err error) {
	if len(args) != 1 {
		return cmd, fmt.Errorf("%s command takes 1 arguments", ty)
	}

	label := args[0]

	// validation
	if err := p.validateSymbol(ty, label); err != nil {
		return cmd, err
	}

	return Command{
		Type: ty,
		Label: &LabelArgs{
			Label: label,
		},
	}, nil
}

func (p *Parser) mapFunctionCommand(ty CommandType, args []string) (cmd Command, err error) {
	if len(args) != 2 {
		return cmd, fmt.Errorf("%s command takes 2 arguments", ty)
	}

	name := args[0]
	numStr := args[1]

	// validations
	if err := p.validateSymbol(ty, name); err != nil {
		return cmd, err
	}

	num, err := strconv.ParseUint(numStr, 10, 64)
	if err != nil {
		return cmd, fmt.Errorf("%s command 2nd arg is must be number", ty)
	}

	return Command{
		Type: ty,
		Function: &FunctionArgs{
			Name: name,
			Num:  num,
		},
	}, nil
}

func (p *Parser) validateSymbol(ty CommandType, sym string) error {
	for i, c := range sym {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c == '.' || c == ':' {
			continue
		}
		if i != 0 && c >= '0' && c <= '9' {
			continue
		}
		return fmt.Errorf("symbol \"%s\": invalid char '%c' at %d", sym, c, i)
	}
	return nil
}

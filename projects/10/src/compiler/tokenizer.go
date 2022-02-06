package compiler

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
}

type Tokens []Token

type TokenType string

const (
	TokenTypeKeyword      TokenType = "keyword"
	TokenTypeSymbol       TokenType = "symbol"
	TokenTypeIntegerConst TokenType = "integerConstant"
	TokenTypeStringConst  TokenType = "stringConstant"
	TokenTypeIdentifier   TokenType = "identifier"
)

var (
	keywords = []string{
		"class",
		"constructor",
		"function",
		"method",
		"field",
		"static",
		"var",
		"int",
		"char",
		"boolean",
		"void",
		"true",
		"false",
		"null",
		"this",
		"let",
		"do",
		"if",
		"else",
		"while",
		"return",
	}

	symbols = []rune{
		'{', '}', '(', ')', '[', ']', '.', ',', ';', '+', '-', '*', '/', '&', '|', ',', '<', '>', '=', '~',
	}

	spaces = []rune{' ', '\t'}
)

func Tokenize(input io.Reader) (Tokens, error) {
	t := newTokenizer(input)
	return t.do()
}

type tokenizer struct {
	scanner      *bufio.Scanner
	line         int
	rangeComment bool
}

func newTokenizer(input io.Reader) tokenizer {
	return tokenizer{
		scanner: bufio.NewScanner(input),
	}
}

func (t *tokenizer) do() ([]Token, error) {
	var ret []Token
	for {
		line, n, err := t.nextLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		tokens, err := t.tokenize(line, n)
		if err != nil {
			return nil, err
		}
		ret = append(ret, tokens...)
	}
	return ret, nil
}

func (t *tokenizer) tokenize(code string, line int) ([]Token, error) {
	var tokens []Token

	for {
		if len(code) == 0 {
			break
		}

		ch := rune(code[0])
		switch {
		case runeInclude(spaces, ch): // space
			code = code[1:]

		case runeInclude(symbols, ch): // symbol
			tokens = append(tokens, Token{Type: TokenTypeSymbol, Value: string(ch), Line: line})
			code = code[1:]

		case ch == '"': // string
			s := tokenStrRegexp.FindString(code)
			if s == "" {
				return nil, fmt.Errorf("string not closed: line %d", line)
			}
			tokens = append(tokens, Token{Type: TokenTypeStringConst, Value: strings.Trim(s, `"`), Line: line})
			code = code[len(s):]

		case '0' <= ch && ch <= '9': // int
			s := tokenIntRegexp.FindString(code)
			tokens = append(tokens, Token{Type: TokenTypeIntegerConst, Value: s, Line: line})
			code = code[len(s):]

		default: // keyword or identifier
			i := tokenIdentRegexp.FindString(code)
			if stringInclude(keywords, i) {
				tokens = append(tokens, Token{Type: TokenTypeKeyword, Value: i, Line: line})
			} else {
				tokens = append(tokens, Token{Type: TokenTypeIdentifier, Value: i, Line: line})
			}

			code = code[len(i):]
		}
	}

	return tokens, nil
}

var (
	tokenStrRegexp   = regexp.MustCompile(`^"[^"]*"`)
	tokenIntRegexp   = regexp.MustCompile("^[0-9]+")
	tokenIdentRegexp = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*")
)

func (t *tokenizer) nextLine() (string, int, error) {
	var line string
	for {
		t.line++
		if !t.scanner.Scan() {
			if t.rangeComment {
				return "", t.line, fmt.Errorf("comment not closed: line %d", t.line)
			}
			return "", t.line, io.EOF
		}

		if err := t.scanner.Err(); err != nil {
			return "", t.line, err
		}

		line = t.scanner.Text()
		line = t.trimComment(line)
		if line == "" {
			continue
		}
		break
	}
	return line, t.line, nil
}

func (t *tokenizer) trimComment(str string) string {

	if t.rangeComment {
		pos := strings.Index(str, "*/")
		if pos == -1 {
			return ""
		}
		str = str[pos+2:]
		t.rangeComment = false
	}

	pos := strings.Index(str, "//")
	if pos != -1 {
		str = str[:pos]
	}

	for {
		pos1 := strings.Index(str, "/*")
		if pos1 == -1 {
			break
		}
		t.rangeComment = true

		pos2 := strings.Index(str, "*/")
		if pos2 == -1 {
			str = str[:pos1]
			break
		}
		str = str[:pos1] + str[pos2+2:]
		t.rangeComment = false
	}
	return strings.TrimSpace(str)
}

func (ts Tokens) ToXML() io.Reader {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("<tokens>\n")
	for _, t := range ts {
		buf.WriteString("<" + string(t.Type) + "> " + t.Value + " </" + string(t.Type) + ">\n")
	}
	buf.WriteString("</tokens>\n")
	return buf
}

func escapeSymbolXML(sym rune) string {
	switch sym {
	case '<':
		return "&lt;"
	case '>':
		return "&gt;"
	case '&':
		return "&amp;"
	case '\'':
		return "&apos;"
	case '"':
		return "&quot;"
	default:
		return string(sym)
	}
}

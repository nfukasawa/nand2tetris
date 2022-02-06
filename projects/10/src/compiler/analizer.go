package compiler

import "fmt"

type Class struct {
	ClassName      string
	ClassVarDecs   []ClassVarDec
	SubRoutineDecs []SubroutineDec
}

type Type string

const (
	TypeInt     Type = "int"
	TypeChar    Type = "char"
	TypeBoolean Type = "boolean"

	TypeVoid Type = "void"
)

type ClassVarDec struct {
	ClassVarDecType ClassVarDecType
	VarType         Type
	VarNames        []string
}

type ClassVarDecType string

const (
	ClassVarDecTypeStatic ClassVarDecType = "static"
	ClassVarDecTypeField  ClassVarDecType = "field"
)

type SubroutineDec struct {
	SubRoutineType SubRoutineType
	RetType        Type
	SubroutineName string
	ParameterList  []Parameter
	SubroutineBody SubroutineBody
}

type SubRoutineType string

const (
	SubRoutineTypeConstructor SubRoutineType = "constructor"
	SubRoutineTypeFunction    SubRoutineType = "function"
	SubRoutineTypeMethod      SubRoutineType = "method"
)

type Parameter struct {
	VarType Type
	VarName string
}

type SubroutineBody struct {
}

type VarDec struct{}

type Statements []Statement

type Statement struct{}

type LetStatement struct{}

type IfStatement struct{}

type WhileStatement struct{}

type DoStatement struct{}

type ReturnStatement struct{}

type Expression struct{}

type Term struct{}

type SubroutineCall struct{}

type ExpressionList struct{}

type KeywordConstant string

const (
	KeywordConstantTrue  KeywordConstant = "true"
	KeywordConstantFalse KeywordConstant = "false"
	KeywordConstantNull  KeywordConstant = "null"
	KeywordConstantThis  KeywordConstant = "this"
)

func Analyze(tokens Tokens) (*Class, error) {
	a := newAnalyzer(tokens)
	return a.parseClass()
}

type analyzer struct {
	tokens Tokens
}

func newAnalyzer(tokens Tokens) analyzer {
	return analyzer{
		tokens: tokens,
	}
}

func (a *analyzer) parseClass() (*Class, error) {
	token := a.popToken()
	if err := assertToken(token, TokenTypeKeyword, "class"); err != nil {
		return nil, err
	}

	token = a.popToken()
	if err := assertToken(token, TokenTypeIdentifier); err != nil {
		return nil, err
	}
	className := token.Value

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "{"); err != nil {
		return nil, err
	}

	var vDecs []ClassVarDec
	for {
		dec, err := a.parseClassVarDec()
		if err != nil {
			return nil, err
		}
		if dec == nil {
			break
		}
		vDecs = append(vDecs, *dec)
	}

	var srDecs []SubroutineDec
	for {
		dec, err := a.parseSubroutineDec()
		if err != nil {
			return nil, err
		}
		if dec == nil {
			break
		}
		srDecs = append(srDecs, *dec)
	}

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "}"); err != nil {
		return nil, err
	}

	return &Class{
		ClassName:      className,
		ClassVarDecs:   vDecs,
		SubRoutineDecs: srDecs,
	}, nil
}

func (a *analyzer) parseType() (Type, error) {
	token := a.popToken()
	switch token.Type {
	case TokenTypeKeyword:
		switch token.Value {
		case "int", "char", "boolean":
			return Type(token.Value), nil
		default:
			return "", fmt.Errorf("int, char, boolean or className is expected, but got %+v", token)
		}
	case TokenTypeIdentifier:
		return Type(token.Value), nil
	default:
		return "", fmt.Errorf("int, char, boolean or className is expected, but got %+v", token)
	}
}

func (a *analyzer) parseRetType() (Type, error) {
	token := a.popToken()
	switch token.Type {
	case TokenTypeKeyword:
		switch token.Value {
		case "int", "char", "boolean", "void":
			return Type(token.Value), nil
		default:
			return "", fmt.Errorf("int, char, boolean, void or className is expected, but got %+v", token)
		}
	case TokenTypeIdentifier:
		return Type(token.Value), nil
	default:
		return "", fmt.Errorf("int, char, boolean, void or className is expected, but got %+v", token)
	}
}

func (a *analyzer) parseClassVarDec() (*ClassVarDec, error) {
	if err := assertToken(a.topToken(), TokenTypeKeyword, "static", "field"); err != nil {
		return nil, nil
	}

	token := a.popToken()
	clsVarTy := ClassVarDecType(token.Value)

	varTy, err := a.parseType()
	if err != nil {
		return nil, err
	}

	var varNames []string
	for {
		token = a.popToken()
		if err := assertToken(token, TokenTypeIdentifier); err != nil {
			return nil, err
		}
		varNames = append(varNames, token.Value)

		token = a.popToken()
		if err := assertToken(token, TokenTypeSymbol, ",", ";"); err != nil {
			return nil, err
		}
		if token.Value == ";" {
			break
		}
	}

	return &ClassVarDec{
		ClassVarDecType: clsVarTy,
		VarType:         varTy,
		VarNames:        varNames,
	}, nil
}

func (a *analyzer) parseSubroutineDec() (*SubroutineDec, error) {
	if err := assertToken(a.topToken(), TokenTypeKeyword, "constructor", "function", "method"); err != nil {
		return nil, nil
	}

	token := a.popToken()
	srTy := SubRoutineType(token.Value)

	retTy, err := a.parseRetType()
	if err != nil {
		return nil, err
	}

	token = a.popToken()
	if err := assertToken(token, TokenTypeIdentifier); err != nil {
		return nil, err
	}
	srName := token.Value

	params, err := a.parseParamterList()
	if err != nil {
		return nil, err
	}

	body, err := a.parseSubroutineBody()
	if err != nil {
		return nil, err
	}

	return &SubroutineDec{
		SubRoutineType: srTy,
		RetType:        retTy,
		SubroutineName: srName,
		ParameterList:  params,
		SubroutineBody: *body,
	}, nil
}

func (a *analyzer) parseParamterList() ([]Parameter, error) {
	token := a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "("); err != nil {
		return nil, err
	}

	token = a.topToken()
	if token.Type == TokenTypeSymbol && token.Value == ")" {
		a.popToken()
		return nil, nil
	}

	var params []Parameter
	for {
		ty, err := a.parseType()
		if err != nil {
			return nil, err
		}

		token = a.popToken()
		if err := assertToken(token, TokenTypeIdentifier); err != nil {
			return nil, err
		}

		params = append(params, Parameter{VarType: ty, VarName: token.Value})

		token = a.popToken()
		if err := assertToken(token, TokenTypeSymbol, ",", ")"); err != nil {
			return nil, err
		}
		if token.Value == ")" {
			break
		}
	}
	return params, nil
}

func (a *analyzer) parseSubroutineBody() (*SubroutineBody, error) {

	// TODO: ignoring body
	depth := 0
	var token *Token
	for {
		token = a.popToken()
		if token == nil {
			break
		}
		switch token.Value {
		case "{":
			depth++
		case "}":
			depth--
		}
		if depth == 0 {
			break
		}
	}

	return &SubroutineBody{}, nil
}

func (a *analyzer) popToken() *Token {
	if len(a.tokens) == 0 {
		return nil
	}
	ret := &a.tokens[0]
	a.tokens = a.tokens[1:]
	return ret
}

func (a *analyzer) topToken() *Token {
	if len(a.tokens) == 0 {
		return nil
	}
	return &a.tokens[0]
}

func assertToken(token *Token, expectedType TokenType, candidateValues ...string) error {
	if token == nil {
		return assertTokenError(token, expectedType, candidateValues)
	}

	if token.Type != expectedType {
		return assertTokenError(token, expectedType, candidateValues)
	}

	if len(candidateValues) != 0 && !stringInclude(candidateValues, token.Value) {
		return assertTokenError(token, expectedType, candidateValues)
	}

	return nil
}

func assertTokenError(token *Token, expectedType TokenType, candidateValues []string) error {
	str := fmt.Sprintf("token type '%s' is expected", expectedType)
	if len(candidateValues) > 0 {
		str += fmt.Sprintf(" with values %+v", candidateValues)
	}
	if token != nil {
		str += fmt.Sprintf(", but got %+v", token)
	}
	return fmt.Errorf(str)
}

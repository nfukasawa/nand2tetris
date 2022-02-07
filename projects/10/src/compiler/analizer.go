package compiler

import (
	"fmt"
	"strconv"
)

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
	VarDecs    []VarDec
	Statements []Statement
}

type VarDec struct {
	VarType  Type
	VarNames []string
}

type Statement struct {
	Type            StatementType
	LetStatement    *LetStatement
	IfStatement     *IfStatement
	WhileStatement  *WhileStatement
	DoStatement     *DoStatement
	ReturnStatement *ReturnStatement
}

type StatementType string

const (
	StatementTypeLet    StatementType = "let"
	StatementTypeIf     StatementType = "if"
	StatementTypeWhile  StatementType = "while"
	StatementTypeDo     StatementType = "do"
	StatementTypeReturn StatementType = "return"
)

type LetStatement struct {
	VarName  string
	Index    *Expression
	VarValue Expression
}

type IfStatement struct {
	Condition      Expression
	IfStatements   []Statement
	ElseStatements []Statement
}

type WhileStatement struct {
	Condition  Expression
	Statements []Statement
}

type DoStatement struct {
	SubroutineCall SubroutineCall
}

type ReturnStatement struct {
	Expression *Expression
}

type Expression struct {
	Term Term
	Tail []ExpressionTail
}
type ExpressionTail struct {
	Op   Op
	Term Term
}

type Term struct {
	Type            TermType
	IntegerConst    *int64
	StringConst     *string
	KeywordConstant *string
	VarName         *string
	Index           *Expression
	SubroutineCall  *SubroutineCall
	Expression      *Expression
	UnaryOp         *UnaryOp
	UnaryOpTerm     *Term
}

type TermType string

const (
	TermTypeIntegerConst   TermType = "integerConstant"
	TermTypeStringConst    TermType = "stringConstant"
	TermTypeKeywordConst   TermType = "keywordConstant"
	TermTypeVarName        TermType = "varName"
	TermTypeVarNameIndex   TermType = "varNameIndex"
	TermTypeSubroutineCall TermType = "subroutineCall"
	TermTypeExpression     TermType = "expression"
	TermTypeUnaryOp        TermType = "unary"
)

type SubroutineCall struct {
	ClassOrVarName *string
	SubroutineName string
	ExpressionList []Expression
}

type Op string

type UnaryOp string

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
	if !checkToken(a.topToken(), TokenTypeKeyword, "static", "field") {
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
	if !checkToken(a.topToken(), TokenTypeKeyword, "constructor", "function", "method") {
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

	token := a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "{"); err != nil {
		return nil, err
	}

	var decs []VarDec
	for {
		dec, err := a.parseVarDec()
		if err != nil {
			return nil, err
		}
		if dec == nil {
			break
		}
		decs = append(decs, *dec)
	}

	statements, err := a.parseStatements()
	if err != nil {
		return nil, err
	}

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "}"); err != nil {
		return nil, err
	}

	return &SubroutineBody{
		VarDecs:    decs,
		Statements: statements,
	}, nil
}

func (a *analyzer) parseVarDec() (*VarDec, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "var") {
		return nil, nil
	}
	a.popToken()

	varTy, err := a.parseType()
	if err != nil {
		return nil, err
	}

	var varNames []string
	for {
		token := a.popToken()
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

	return &VarDec{
		VarType:  varTy,
		VarNames: varNames,
	}, nil
}

func (a *analyzer) parseStatements() ([]Statement, error) {
	var statements []Statement
	for {
		statement, err := a.parseStatement()
		if err != nil {
			return nil, err
		}
		if statement == nil {
			break
		}
		statements = append(statements, *statement)
	}
	return statements, nil
}

func (a *analyzer) parseStatement() (*Statement, error) {
	if checkToken(a.topToken(), TokenTypeSymbol, "}") {
		return nil, nil
	}

	{
		s, err := a.parseLetStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			return &Statement{Type: StatementTypeLet, LetStatement: s}, nil
		}
	}

	{
		s, err := a.parseIfStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			return &Statement{Type: StatementTypeIf, IfStatement: s}, nil
		}
	}

	{
		s, err := a.parseWhileStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			return &Statement{Type: StatementTypeWhile, WhileStatement: s}, nil
		}
	}

	{
		s, err := a.parseDoStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			return &Statement{Type: StatementTypeDo, DoStatement: s}, nil
		}
	}

	{
		s, err := a.parseReturnStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			return &Statement{Type: StatementTypeReturn, ReturnStatement: s}, nil
		}
	}

	return nil, fmt.Errorf("invalid statement: %+v", a.topToken())
}

func (a *analyzer) parseLetStatement() (*LetStatement, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "let") {
		return nil, nil
	}
	a.popToken()

	token := a.popToken()
	if err := assertToken(token, TokenTypeIdentifier); err != nil {
		return nil, err
	}
	varName := token.Value

	var idx *Expression
	var err error
	if checkToken(a.topToken(), TokenTypeSymbol, "[") {
		a.popToken()

		idx, err = a.parseExpression()
		if err != nil {
			return nil, err
		}

		if err := assertToken(a.popToken(), TokenTypeSymbol, "]"); err != nil {
			return nil, err
		}
	}

	if err := assertToken(a.popToken(), TokenTypeSymbol, "="); err != nil {
		return nil, err
	}

	val, err := a.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := assertToken(a.popToken(), TokenTypeSymbol, ";"); err != nil {
		return nil, err
	}

	return &LetStatement{
		VarName:  varName,
		Index:    idx,
		VarValue: *val,
	}, nil
}

func (a *analyzer) parseIfStatement() (*IfStatement, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "if") {
		return nil, nil
	}
	a.popToken()

	if err := assertToken(a.popToken(), TokenTypeSymbol, "("); err != nil {
		return nil, err
	}
	cond, err := a.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := assertToken(a.popToken(), TokenTypeSymbol, ")"); err != nil {
		return nil, err
	}

	if err := assertToken(a.popToken(), TokenTypeSymbol, "{"); err != nil {
		return nil, err
	}
	statements, err := a.parseStatements()
	if err != nil {
		return nil, err
	}
	if err := assertToken(a.popToken(), TokenTypeSymbol, "}"); err != nil {
		return nil, err
	}

	var elseStatements []Statement
	if checkToken(a.topToken(), TokenTypeKeyword, "else") {
		a.popToken()
		if err := assertToken(a.popToken(), TokenTypeSymbol, "{"); err != nil {
			return nil, err
		}
		elseStatements, err = a.parseStatements()
		if err != nil {
			return nil, err
		}
		if err := assertToken(a.popToken(), TokenTypeSymbol, "}"); err != nil {
			return nil, err
		}
	}

	return &IfStatement{
		Condition:      *cond,
		IfStatements:   statements,
		ElseStatements: elseStatements,
	}, nil
}

func (a *analyzer) parseWhileStatement() (*WhileStatement, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "while") {
		return nil, nil
	}
	a.popToken()

	if err := assertToken(a.popToken(), TokenTypeSymbol, "("); err != nil {
		return nil, err
	}
	cond, err := a.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := assertToken(a.popToken(), TokenTypeSymbol, ")"); err != nil {
		return nil, err
	}

	if err := assertToken(a.popToken(), TokenTypeSymbol, "{"); err != nil {
		return nil, err
	}
	statements, err := a.parseStatements()
	if err != nil {
		return nil, err
	}
	if err := assertToken(a.popToken(), TokenTypeSymbol, "}"); err != nil {
		return nil, err
	}

	return &WhileStatement{
		Condition:  *cond,
		Statements: statements,
	}, nil
}

func (a *analyzer) parseDoStatement() (*DoStatement, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "do") {
		return nil, nil
	}
	a.popToken()

	call, err := a.parseSubroutineCall()
	if err != nil {
		return nil, err
	}

	if err := assertToken(a.popToken(), TokenTypeSymbol, ";"); err != nil {
		return nil, err
	}

	return &DoStatement{
		SubroutineCall: *call,
	}, nil
}

func (a *analyzer) parseReturnStatement() (*ReturnStatement, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "return") {
		return nil, nil
	}
	a.popToken()

	if checkToken(a.topToken(), TokenTypeSymbol, ";") {
		a.popToken()
		return &ReturnStatement{}, nil
	}

	exp, err := a.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := assertToken(a.popToken(), TokenTypeSymbol, ";"); err != nil {
		return nil, err
	}

	return &ReturnStatement{
		Expression: exp,
	}, nil
}

func (a *analyzer) parseExpression() (*Expression, error) {
	term, err := a.parseTerm()
	if err != nil {
		return nil, err
	}

	var tail []ExpressionTail
	for {
		token := a.topToken()
		if checkToken(token, TokenTypeSymbol, "+", "-", "*", "/", "&", "|", "<", ">", "=") {
			op := Op(token.Value)
			a.popToken()
			term, err := a.parseTerm()
			if err != nil {
				return nil, err
			}
			tail = append(tail, ExpressionTail{Op: op, Term: *term})
			continue
		}
		break
	}

	return &Expression{Term: *term, Tail: tail}, nil
}

func (a *analyzer) parseTerm() (*Term, error) {

	token := a.popToken()
	switch {
	case checkToken(token, TokenTypeIntegerConst):
		i, err := strconv.ParseInt(token.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("integerConstant parse error: %+v", token)
		}
		return &Term{Type: TermTypeIntegerConst, IntegerConst: &i}, nil

	case checkToken(token, TokenTypeStringConst):
		return &Term{Type: TermTypeStringConst, StringConst: &token.Value}, nil

	case checkToken(token, TokenTypeKeyword, "true", "false", "null", "this"):
		return &Term{Type: TermTypeKeywordConst, KeywordConstant: &token.Value}, nil

	case checkToken(token, TokenTypeIdentifier):
		next := a.topToken()
		switch {
		case checkToken(next, TokenTypeSymbol, "(", "."):
			a.pushToken(*token)
			call, err := a.parseSubroutineCall()
			if err != nil {
				return nil, err
			}
			return &Term{Type: TermTypeSubroutineCall, SubroutineCall: call}, nil

		case checkToken(next, TokenTypeSymbol, "["):
			exp, err := a.parseExpression()
			if err != nil {
				return nil, err
			}
			return &Term{Type: TermTypeVarNameIndex, VarName: &token.Value, Index: exp}, nil

		default:
			return &Term{Type: TermTypeVarName, VarName: &token.Value}, nil
		}

	case checkToken(token, TokenTypeSymbol, "("):
		exp, err := a.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := assertToken(a.popToken(), TokenTypeSymbol, ")"); err != nil {
			return nil, err
		}
		return &Term{Type: TermTypeExpression, Expression: exp}, nil

	case checkToken(token, TokenTypeSymbol, "-", "~"):
		term, err := a.parseTerm()
		if err != nil {
			return nil, err
		}
		op := UnaryOp(token.Value)
		return &Term{
			Type:        TermTypeUnaryOp,
			UnaryOp:     &op,
			UnaryOpTerm: term}, nil
	default:
		return nil, fmt.Errorf("term is expected: %+v", token)
	}
}

func (a *analyzer) parseSubroutineCall() (*SubroutineCall, error) {
	token := a.popToken()
	if err := assertToken(token, TokenTypeIdentifier); err != nil {
		return nil, err
	}

	next := a.topToken()
	if err := assertToken(next, TokenTypeSymbol, "(", "."); err != nil {
		return nil, err
	}

	var classOrVarName *string
	var subroutineName string
	switch next.Value {
	case "(":
		subroutineName = token.Value
	case ".":
		a.popToken()
		classOrVarName = &token.Value
		token := a.popToken()
		if err := assertToken(token, TokenTypeIdentifier); err != nil {
			return nil, err
		}
	}

	exps, err := a.parseExpressionList()
	if err != nil {
		return nil, err
	}

	return &SubroutineCall{
		ClassOrVarName: classOrVarName,
		SubroutineName: subroutineName,
		ExpressionList: exps,
	}, nil
}

func (a *analyzer) parseExpressionList() ([]Expression, error) {
	token := a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "("); err != nil {
		return nil, err
	}

	token = a.topToken()
	if token.Type == TokenTypeSymbol && token.Value == ")" {
		a.popToken()
		return nil, nil
	}

	var exps []Expression
	for {
		exp, err := a.parseExpression()
		if err != nil {
			return nil, err
		}

		exps = append(exps, *exp)

		token = a.popToken()
		if err := assertToken(token, TokenTypeSymbol, ",", ")"); err != nil {
			return nil, err
		}
		if token.Value == ")" {
			break
		}
	}
	return exps, nil
}

func (a *analyzer) topToken() *Token {
	if len(a.tokens) == 0 {
		return nil
	}
	return &a.tokens[0]
}

func (a *analyzer) popToken() *Token {
	if len(a.tokens) == 0 {
		return nil
	}
	ret := &a.tokens[0]
	a.tokens = a.tokens[1:]
	return ret
}

func (a *analyzer) pushToken(token Token) {
	a.tokens = append(Tokens{token}, a.tokens...)
}

func checkToken(token *Token, expectedType TokenType, candidateValues ...string) bool {
	if token == nil {
		return false
	}

	if token.Type != expectedType {
		return false
	}

	if len(candidateValues) != 0 && !stringInclude(candidateValues, token.Value) {
		return false
	}

	return true
}

func assertToken(token *Token, expectedType TokenType, candidateValues ...string) error {
	if !checkToken(token, expectedType, candidateValues...) {
		str := fmt.Sprintf("token type '%s' is expected", expectedType)
		if len(candidateValues) > 0 {
			str += fmt.Sprintf(" with values %+v", candidateValues)
		}
		if token != nil {
			str += fmt.Sprintf(", but got %+v", token)
		}
		return fmt.Errorf(str)
	}
	return nil
}

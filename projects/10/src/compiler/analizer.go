package compiler

import (
	"fmt"
	"strconv"
)

type Class struct {
	ClassName      string
	ClassVarDecs   []ClassVarDec
	SubRoutineDecs []SubroutineDec

	XML XMLElm
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

	XML XMLElm
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
	ParameterList  ParameterList
	SubroutineBody SubroutineBody

	XML XMLElm
}

type SubRoutineType string

const (
	SubRoutineTypeConstructor SubRoutineType = "constructor"
	SubRoutineTypeFunction    SubRoutineType = "function"
	SubRoutineTypeMethod      SubRoutineType = "method"
)

type ParameterList struct {
	Paramters []Parameter

	XML XMLElm
}

type Parameter struct {
	VarType Type
	VarName string

	XML XMLElm
}

type SubroutineBody struct {
	VarDecs    []VarDec
	Statements Statements

	XML XMLElm
}

type VarDec struct {
	VarType  Type
	VarNames []string

	XML XMLElm
}

type Statements struct {
	Statements []Statement

	XML XMLElm
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

	XML XMLElm
}

type IfStatement struct {
	Condition      Expression
	IfStatements   Statements
	ElseStatements Statements

	XML XMLElm
}

type WhileStatement struct {
	Condition  Expression
	Statements Statements

	XML XMLElm
}

type DoStatement struct {
	SubroutineCall SubroutineCall

	XML XMLElm
}

type ReturnStatement struct {
	Expression *Expression

	XML XMLElm
}

type Expression struct {
	Term Term
	Tail []ExpressionTail

	XML XMLElm
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

	XML XMLElm
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
	ExpressionList ExpressionList

	XML XMLElm
}

type ExpressionList struct {
	Expressions []Expression

	XML XMLElm
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
	cls := Class{XML: XMLElm{Name: "class"}}

	token := a.popToken()
	if err := assertToken(token, TokenTypeKeyword, "class"); err != nil {
		return nil, err
	}
	cls.XML.AddChild(token)

	token = a.popToken()
	if err := assertToken(token, TokenTypeIdentifier); err != nil {
		return nil, err
	}
	cls.ClassName = token.Value
	cls.XML.AddChild(token)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "{"); err != nil {
		return nil, err
	}
	cls.XML.AddChild(token)

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
		cls.XML.AddChild(dec)
	}
	cls.ClassVarDecs = vDecs

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
		cls.XML.AddChild(dec)
	}
	cls.SubRoutineDecs = srDecs

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "}"); err != nil {
		return nil, err
	}
	cls.XML.AddChild(token)

	return &cls, nil
}

func (a *analyzer) parseType() (Type, *Token, error) {
	token := a.popToken()
	switch token.Type {
	case TokenTypeKeyword:
		switch token.Value {
		case "int", "char", "boolean":
			return Type(token.Value), token, nil
		default:
			return "", nil, fmt.Errorf("int, char, boolean or className is expected, but got %+v", token)
		}
	case TokenTypeIdentifier:
		return Type(token.Value), token, nil
	default:
		return "", nil, fmt.Errorf("int, char, boolean or className is expected, but got %+v", token)
	}
}

func (a *analyzer) parseRetType() (Type, *Token, error) {
	token := a.popToken()
	switch token.Type {
	case TokenTypeKeyword:
		switch token.Value {
		case "int", "char", "boolean", "void":
			return Type(token.Value), token, nil
		default:
			return "", nil, fmt.Errorf("int, char, boolean, void or className is expected, but got %+v", token)
		}
	case TokenTypeIdentifier:
		return Type(token.Value), token, nil
	default:
		return "", nil, fmt.Errorf("int, char, boolean, void or className is expected, but got %+v", token)
	}
}

func (a *analyzer) parseClassVarDec() (*ClassVarDec, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "static", "field") {
		return nil, nil
	}
	dec := ClassVarDec{XML: XMLElm{Name: "classVarDec"}}

	token := a.popToken()
	dec.ClassVarDecType = ClassVarDecType(token.Value)
	dec.XML.AddChild(token)

	var err error
	dec.VarType, token, err = a.parseType()
	if err != nil {
		return nil, err
	}
	dec.XML.AddChild(token)

	var varNames []string
	for {
		token = a.popToken()
		if err := assertToken(token, TokenTypeIdentifier); err != nil {
			return nil, err
		}
		varNames = append(varNames, token.Value)
		dec.XML.AddChild(token)

		token = a.popToken()
		if err := assertToken(token, TokenTypeSymbol, ",", ";"); err != nil {
			return nil, err
		}
		dec.XML.AddChild(token)
		if token.Value == ";" {
			break
		}
	}
	dec.VarNames = varNames

	return &dec, nil
}

func (a *analyzer) parseSubroutineDec() (*SubroutineDec, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "constructor", "function", "method") {
		return nil, nil
	}

	dec := SubroutineDec{XML: XMLElm{Name: "subroutineDec"}}
	token := a.popToken()
	dec.SubRoutineType = SubRoutineType(token.Value)
	dec.XML.AddChild(token)

	var err error
	dec.RetType, token, err = a.parseRetType()
	if err != nil {
		return nil, err
	}
	dec.XML.AddChild(token)

	token = a.popToken()
	if err := assertToken(token, TokenTypeIdentifier); err != nil {
		return nil, err
	}
	dec.SubroutineName = token.Value
	dec.XML.AddChild(token)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "("); err != nil {
		return nil, err
	}
	dec.XML.AddChild(token)
	params, err := a.parseParamterList()
	if err != nil {
		return nil, err
	}
	dec.XML.AddChild(params)
	dec.XML.AddChild(a.popToken()) // ")"

	body, err := a.parseSubroutineBody()
	if err != nil {
		return nil, err
	}
	dec.SubroutineBody = *body
	dec.XML.AddChild(body)

	return &dec, nil
}

func (a *analyzer) parseParamterList() (*ParameterList, error) {

	params := &ParameterList{XML: XMLElm{Name: "parameterList", Children: []XMLElm{}}}

	token := a.topToken()
	if token.Type == TokenTypeSymbol && token.Value == ")" {
		return params, nil
	}

	for {
		ty, token, err := a.parseType()
		if err != nil {
			return nil, err
		}
		params.XML.AddChild(token)

		token = a.popToken()
		if err := assertToken(token, TokenTypeIdentifier); err != nil {
			return nil, err
		}
		params.XML.AddChild(token)

		params.Paramters = append(params.Paramters, Parameter{VarType: ty, VarName: token.Value})

		token = a.topToken()
		if err := assertToken(token, TokenTypeSymbol, ",", ")"); err != nil {
			return nil, err
		}
		if token.Value == "," {
			params.XML.AddChild(a.popToken())
			continue
		}
		break
	}
	return params, nil
}

func (a *analyzer) parseSubroutineBody() (*SubroutineBody, error) {

	body := SubroutineBody{XML: XMLElm{Name: "subroutineBody"}}

	token := a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "{"); err != nil {
		return nil, err
	}
	body.XML.AddChild(token)

	for {
		dec, err := a.parseVarDec()
		if err != nil {
			return nil, err
		}
		if dec == nil {
			break
		}
		body.VarDecs = append(body.VarDecs, *dec)
		body.XML.AddChild(dec)
	}

	var err error
	statements, err := a.parseStatements()
	if err != nil {
		return nil, err
	}
	body.Statements = *statements
	body.XML.AddChild(statements)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "}"); err != nil {
		return nil, err
	}
	body.XML.AddChild(token)

	return &body, nil
}

func (a *analyzer) parseVarDec() (*VarDec, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "var") {
		return nil, nil
	}

	dec := VarDec{XML: XMLElm{Name: "varDec"}}
	dec.XML.AddChild(a.popToken())

	varTy, token, err := a.parseType()
	if err != nil {
		return nil, err
	}
	dec.VarType = varTy
	dec.XML.AddChild(token)

	for {
		token := a.popToken()
		if err := assertToken(token, TokenTypeIdentifier); err != nil {
			return nil, err
		}
		dec.XML.AddChild(token)

		dec.VarNames = append(dec.VarNames, token.Value)

		token = a.popToken()
		if err := assertToken(token, TokenTypeSymbol, ",", ";"); err != nil {
			return nil, err
		}
		dec.XML.AddChild(token)

		if token.Value == ";" {
			break
		}
	}

	return &dec, nil
}

func (a *analyzer) parseStatements() (*Statements, error) {
	statements := Statements{XML: XMLElm{Name: "statements", Children: []XMLElm{}}}
	for {
		statement, err := a.parseStatement()
		if err != nil {
			return nil, err
		}
		if statement == nil {
			break
		}
		statements.Statements = append(statements.Statements, *statement)
		statements.XML.AddChild(statement)
	}
	return &statements, nil
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
	statement := LetStatement{XML: XMLElm{Name: "letStatement"}}
	statement.XML.AddChild(a.popToken())

	token := a.popToken()
	if err := assertToken(token, TokenTypeIdentifier); err != nil {
		return nil, err
	}
	statement.VarName = token.Value
	statement.XML.AddChild(token)

	var err error
	if checkToken(a.topToken(), TokenTypeSymbol, "[") {
		statement.XML.AddChild(a.popToken())

		statement.Index, err = a.parseExpression()
		if err != nil {
			return nil, err
		}
		statement.XML.AddChild(statement.Index)

		token := a.popToken()
		if err := assertToken(token, TokenTypeSymbol, "]"); err != nil {
			return nil, err
		}
		statement.XML.AddChild(token)
	}

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "="); err != nil {
		return nil, err
	}
	statement.XML.AddChild(token)

	val, err := a.parseExpression()
	if err != nil {
		return nil, err
	}
	statement.VarValue = *val
	statement.XML.AddChild(val)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, ";"); err != nil {
		return nil, err
	}
	statement.XML.AddChild(token)

	return &statement, nil
}

func (a *analyzer) parseIfStatement() (*IfStatement, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "if") {
		return nil, nil
	}
	statement := IfStatement{XML: XMLElm{Name: "ifStatement"}}
	statement.XML.AddChild(a.popToken())

	token := a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "("); err != nil {
		return nil, err
	}
	statement.XML.AddChild(token)

	cond, err := a.parseExpression()
	if err != nil {
		return nil, err
	}
	statement.Condition = *cond
	statement.XML.AddChild(cond)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, ")"); err != nil {
		return nil, err
	}
	statement.XML.AddChild(token)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "{"); err != nil {
		return nil, err
	}
	statement.XML.AddChild(token)

	statements, err := a.parseStatements()
	if err != nil {
		return nil, err
	}
	statement.IfStatements = *statements
	statement.XML.AddChild(statements)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "}"); err != nil {
		return nil, err
	}
	statement.XML.AddChild(token)

	var elseStatements *Statements
	if checkToken(a.topToken(), TokenTypeKeyword, "else") {
		statement.XML.AddChild(a.popToken())

		token = a.popToken()
		if err := assertToken(token, TokenTypeSymbol, "{"); err != nil {
			return nil, err
		}
		statement.XML.AddChild(token)

		elseStatements, err = a.parseStatements()
		if err != nil {
			return nil, err
		}
		statement.ElseStatements = *elseStatements
		statement.XML.AddChild(elseStatements)

		token := a.popToken()
		if err := assertToken(token, TokenTypeSymbol, "}"); err != nil {
			return nil, err
		}
		statement.XML.AddChild(token)
	}

	return &statement, nil
}

func (a *analyzer) parseWhileStatement() (*WhileStatement, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "while") {
		return nil, nil
	}
	statement := WhileStatement{XML: XMLElm{Name: "whileStatement"}}
	statement.XML.AddChild(a.popToken())

	token := a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "("); err != nil {
		return nil, err
	}
	statement.XML.AddChild(token)

	cond, err := a.parseExpression()
	if err != nil {
		return nil, err
	}
	statement.Condition = *cond
	statement.XML.AddChild(cond)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, ")"); err != nil {
		return nil, err
	}
	statement.XML.AddChild(token)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "{"); err != nil {
		return nil, err
	}
	statement.XML.AddChild(token)

	statements, err := a.parseStatements()
	if err != nil {
		return nil, err
	}
	statement.Statements = *statements
	statement.XML.AddChild(statements)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "}"); err != nil {
		return nil, err
	}
	statement.XML.AddChild(token)

	return &statement, nil
}

func (a *analyzer) parseDoStatement() (*DoStatement, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "do") {
		return nil, nil
	}
	statement := DoStatement{XML: XMLElm{Name: "doStatement"}}
	statement.XML.AddChild(a.popToken())

	call, err := a.parseSubroutineCall()
	if err != nil {
		return nil, err
	}
	statement.SubroutineCall = *call
	statement.XML.AddChild(call)

	token := a.popToken()
	if err := assertToken(token, TokenTypeSymbol, ";"); err != nil {
		return nil, err
	}
	statement.XML.AddChild(token)

	return &statement, nil
}

func (a *analyzer) parseReturnStatement() (*ReturnStatement, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "return") {
		return nil, nil
	}
	statement := ReturnStatement{XML: XMLElm{Name: "returnStatement"}}
	statement.XML.AddChild(a.popToken())

	if checkToken(a.topToken(), TokenTypeSymbol, ";") {
		statement.XML.AddChild(a.popToken())
		return &statement, nil
	}

	exp, err := a.parseExpression()
	if err != nil {
		return nil, err
	}
	statement.Expression = exp
	statement.XML.AddChild(exp)

	token := a.popToken()
	if err := assertToken(token, TokenTypeSymbol, ";"); err != nil {
		return nil, err
	}
	statement.XML.AddChild(token)

	return &statement, nil
}

func (a *analyzer) parseExpression() (*Expression, error) {

	exp := Expression{XML: XMLElm{Name: "expression"}}

	term, err := a.parseTerm()
	if err != nil {
		return nil, err
	}
	exp.Term = *term
	exp.XML.AddChild(term)

	for {
		token := a.topToken()
		if checkToken(token, TokenTypeSymbol, "+", "-", "*", "/", "&", "|", "<", ">", "=") {
			op := Op(token.Value)
			exp.XML.AddChild(a.popToken())

			term, err := a.parseTerm()
			if err != nil {
				return nil, err
			}
			exp.XML.AddChild(term)

			exp.Tail = append(exp.Tail, ExpressionTail{Op: op, Term: *term})
			continue
		}
		break
	}

	return &exp, nil
}

func (a *analyzer) parseTerm() (*Term, error) {

	xml := XMLElm{Name: "term"}

	token := a.popToken()
	switch {
	case checkToken(token, TokenTypeIntegerConst):
		i, err := strconv.ParseInt(token.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("integerConstant parse error: %+v", token)
		}
		xml.AddChild(token)
		return &Term{Type: TermTypeIntegerConst, IntegerConst: &i, XML: xml}, nil

	case checkToken(token, TokenTypeStringConst):
		xml.AddChild(token)
		return &Term{Type: TermTypeStringConst, StringConst: &token.Value, XML: xml}, nil

	case checkToken(token, TokenTypeKeyword, "true", "false", "null", "this"):
		xml.AddChild(token)
		return &Term{Type: TermTypeKeywordConst, KeywordConstant: &token.Value, XML: xml}, nil

	case checkToken(token, TokenTypeIdentifier):
		next := a.topToken()
		switch {
		case checkToken(next, TokenTypeSymbol, "(", "."):
			a.pushToken(*token)
			call, err := a.parseSubroutineCall()
			if err != nil {
				return nil, err
			}
			xml.AddChild(call)
			return &Term{Type: TermTypeSubroutineCall, SubroutineCall: call, XML: xml}, nil

		case checkToken(next, TokenTypeSymbol, "["):
			xml.AddChild(token)
			xml.AddChild(a.popToken())
			exp, err := a.parseExpression()
			if err != nil {
				return nil, err
			}
			xml.AddChild(exp)
			token := a.popToken()
			if err := assertToken(token, TokenTypeSymbol, "]"); err != nil {
				return nil, err
			}
			xml.AddChild(token)
			return &Term{Type: TermTypeVarNameIndex, VarName: &token.Value, Index: exp, XML: xml}, nil

		default:
			xml.AddChild(token)
			return &Term{Type: TermTypeVarName, VarName: &token.Value, XML: xml}, nil
		}

	case checkToken(token, TokenTypeSymbol, "("):
		xml.AddChild(token)
		exp, err := a.parseExpression()
		if err != nil {
			return nil, err
		}
		xml.AddChild(exp)
		token := a.popToken()
		if err := assertToken(token, TokenTypeSymbol, ")"); err != nil {
			return nil, err
		}
		xml.AddChild(token)
		return &Term{Type: TermTypeExpression, Expression: exp, XML: xml}, nil

	case checkToken(token, TokenTypeSymbol, "-", "~"):
		xml.AddChild(token)

		op := UnaryOp(token.Value)
		term, err := a.parseTerm()
		if err != nil {
			return nil, err
		}
		xml.AddChild(term)
		return &Term{Type: TermTypeUnaryOp, UnaryOp: &op, UnaryOpTerm: term, XML: xml}, nil
	default:
		return nil, fmt.Errorf("term is expected: %+v", token)
	}
}

func (a *analyzer) parseSubroutineCall() (*SubroutineCall, error) {
	token := a.popToken()
	if err := assertToken(token, TokenTypeIdentifier); err != nil {
		return nil, err
	}

	call := SubroutineCall{XML: XMLElm{SkipLayer: true}}
	call.XML.AddChild(token)

	next := a.topToken()
	if err := assertToken(next, TokenTypeSymbol, "(", "."); err != nil {
		return nil, err
	}

	switch next.Value {
	case "(":
		call.SubroutineName = token.Value
	case ".":
		call.XML.AddChild(a.popToken())
		call.ClassOrVarName = &token.Value

		token := a.popToken()
		if err := assertToken(token, TokenTypeIdentifier); err != nil {
			return nil, err
		}
		call.XML.AddChild(token)
	}

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "("); err != nil {
		return nil, err
	}
	call.XML.AddChild(token)

	exps, err := a.parseExpressionList()
	if err != nil {
		return nil, err
	}
	call.ExpressionList = *exps
	call.XML.AddChild(exps)
	call.XML.AddChild(a.popToken()) // ")"

	return &call, nil
}

func (a *analyzer) parseExpressionList() (*ExpressionList, error) {
	exps := ExpressionList{XML: XMLElm{Name: "expressionList", Children: []XMLElm{}}}
	if checkToken(a.topToken(), TokenTypeSymbol, ")") {
		return &exps, nil
	}

	for {
		exp, err := a.parseExpression()
		if err != nil {
			return nil, err
		}
		exps.XML.AddChild(exp)

		exps.Expressions = append(exps.Expressions, *exp)

		token := a.topToken()
		if err := assertToken(token, TokenTypeSymbol, ",", ")"); err != nil {
			return nil, err
		}
		if token.Value == "," {
			exps.XML.AddChild(a.popToken())
			continue
		}
		break

	}
	return &exps, nil
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

func (x *Class) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *ClassVarDec) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *SubroutineDec) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *ParameterList) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *Parameter) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *SubroutineBody) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *VarDec) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *Statements) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *Statement) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	switch x.Type {
	case StatementTypeLet:
		return x.LetStatement.ToXML()
	case StatementTypeIf:
		return x.IfStatement.ToXML()
	case StatementTypeWhile:
		return x.WhileStatement.ToXML()
	case StatementTypeDo:
		return x.DoStatement.ToXML()
	case StatementTypeReturn:
		return x.ReturnStatement.ToXML()
	default:
		return nil
	}
}
func (x *LetStatement) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *IfStatement) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *WhileStatement) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *DoStatement) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *ReturnStatement) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *Expression) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *Term) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *SubroutineCall) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}
func (x *ExpressionList) ToXML() *XMLElm {
	if x == nil {
		return nil
	}
	return &x.XML
}

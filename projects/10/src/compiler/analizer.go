package compiler

import (
	"fmt"
	"strconv"
)

type Class struct {
	ClassName      string
	ClassVarDecs   []ClassVarDec
	SubRoutineDecs []SubroutineDec

	Node Node
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

	Node Node
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

	Node Node
}

type SubRoutineType string

const (
	SubRoutineTypeConstructor SubRoutineType = "constructor"
	SubRoutineTypeFunction    SubRoutineType = "function"
	SubRoutineTypeMethod      SubRoutineType = "method"
)

type ParameterList struct {
	Paramters []Parameter

	Node Node
}

type Parameter struct {
	VarType Type
	VarName string

	Node Node
}

type SubroutineBody struct {
	VarDecs    []VarDec
	Statements Statements

	Node Node
}

type VarDec struct {
	VarType  Type
	VarNames []string

	Node Node
}

type Statements struct {
	Statements []Statement

	Node Node
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

	Node Node
}

type IfStatement struct {
	Condition      Expression
	IfStatements   Statements
	ElseStatements Statements

	Node Node
}

type WhileStatement struct {
	Condition  Expression
	Statements Statements

	Node Node
}

type DoStatement struct {
	SubroutineCall SubroutineCall

	Node Node
}

type ReturnStatement struct {
	Expression *Expression

	Node Node
}

type Expression struct {
	Term Term
	Tail []ExpressionTail

	Node Node
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

	Node Node
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

	Node Node
}

type ExpressionList struct {
	Expressions []Expression

	Node Node
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
	cls := Class{Node: Node{Name: "class"}}

	token := a.popToken()
	if err := assertToken(token, TokenTypeKeyword, "class"); err != nil {
		return nil, err
	}
	cls.Node.AddChild(token)

	token = a.popToken()
	if err := assertToken(token, TokenTypeIdentifier); err != nil {
		return nil, err
	}
	cls.ClassName = token.Value
	cls.Node.AddChild(token)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "{"); err != nil {
		return nil, err
	}
	cls.Node.AddChild(token)

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
		cls.Node.AddChild(dec)
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
		cls.Node.AddChild(dec)
	}
	cls.SubRoutineDecs = srDecs

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "}"); err != nil {
		return nil, err
	}
	cls.Node.AddChild(token)

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
	dec := ClassVarDec{Node: Node{Name: "classVarDec"}}

	token := a.popToken()
	dec.ClassVarDecType = ClassVarDecType(token.Value)
	dec.Node.AddChild(token)

	var err error
	dec.VarType, token, err = a.parseType()
	if err != nil {
		return nil, err
	}
	dec.Node.AddChild(token)

	var varNames []string
	for {
		token = a.popToken()
		if err := assertToken(token, TokenTypeIdentifier); err != nil {
			return nil, err
		}
		varNames = append(varNames, token.Value)
		dec.Node.AddChild(token)

		token = a.popToken()
		if err := assertToken(token, TokenTypeSymbol, ",", ";"); err != nil {
			return nil, err
		}
		dec.Node.AddChild(token)
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

	dec := SubroutineDec{Node: Node{Name: "subroutineDec"}}
	token := a.popToken()
	dec.SubRoutineType = SubRoutineType(token.Value)
	dec.Node.AddChild(token)

	var err error
	dec.RetType, token, err = a.parseRetType()
	if err != nil {
		return nil, err
	}
	dec.Node.AddChild(token)

	token = a.popToken()
	if err := assertToken(token, TokenTypeIdentifier); err != nil {
		return nil, err
	}
	dec.SubroutineName = token.Value
	dec.Node.AddChild(token)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "("); err != nil {
		return nil, err
	}
	dec.Node.AddChild(token)
	params, err := a.parseParamterList()
	if err != nil {
		return nil, err
	}
	dec.Node.AddChild(params)
	dec.Node.AddChild(a.popToken()) // ")"

	body, err := a.parseSubroutineBody()
	if err != nil {
		return nil, err
	}
	dec.SubroutineBody = *body
	dec.Node.AddChild(body)

	return &dec, nil
}

func (a *analyzer) parseParamterList() (*ParameterList, error) {

	params := &ParameterList{Node: Node{Name: "parameterList", Children: []Node{}}}

	token := a.topToken()
	if token.Type == TokenTypeSymbol && token.Value == ")" {
		return params, nil
	}

	for {
		ty, token, err := a.parseType()
		if err != nil {
			return nil, err
		}
		params.Node.AddChild(token)

		token = a.popToken()
		if err := assertToken(token, TokenTypeIdentifier); err != nil {
			return nil, err
		}
		params.Node.AddChild(token)

		params.Paramters = append(params.Paramters, Parameter{VarType: ty, VarName: token.Value})

		token = a.topToken()
		if err := assertToken(token, TokenTypeSymbol, ",", ")"); err != nil {
			return nil, err
		}
		if token.Value == "," {
			params.Node.AddChild(a.popToken())
			continue
		}
		break
	}
	return params, nil
}

func (a *analyzer) parseSubroutineBody() (*SubroutineBody, error) {

	body := SubroutineBody{Node: Node{Name: "subroutineBody"}}

	token := a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "{"); err != nil {
		return nil, err
	}
	body.Node.AddChild(token)

	for {
		dec, err := a.parseVarDec()
		if err != nil {
			return nil, err
		}
		if dec == nil {
			break
		}
		body.VarDecs = append(body.VarDecs, *dec)
		body.Node.AddChild(dec)
	}

	var err error
	statements, err := a.parseStatements()
	if err != nil {
		return nil, err
	}
	body.Statements = *statements
	body.Node.AddChild(statements)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "}"); err != nil {
		return nil, err
	}
	body.Node.AddChild(token)

	return &body, nil
}

func (a *analyzer) parseVarDec() (*VarDec, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "var") {
		return nil, nil
	}

	dec := VarDec{Node: Node{Name: "varDec"}}
	dec.Node.AddChild(a.popToken())

	varTy, token, err := a.parseType()
	if err != nil {
		return nil, err
	}
	dec.VarType = varTy
	dec.Node.AddChild(token)

	for {
		token := a.popToken()
		if err := assertToken(token, TokenTypeIdentifier); err != nil {
			return nil, err
		}
		dec.Node.AddChild(token)

		dec.VarNames = append(dec.VarNames, token.Value)

		token = a.popToken()
		if err := assertToken(token, TokenTypeSymbol, ",", ";"); err != nil {
			return nil, err
		}
		dec.Node.AddChild(token)

		if token.Value == ";" {
			break
		}
	}

	return &dec, nil
}

func (a *analyzer) parseStatements() (*Statements, error) {
	statements := Statements{Node: Node{Name: "statements", Children: []Node{}}}
	for {
		statement, err := a.parseStatement()
		if err != nil {
			return nil, err
		}
		if statement == nil {
			break
		}
		statements.Statements = append(statements.Statements, *statement)
		statements.Node.AddChild(statement)
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
	statement := LetStatement{Node: Node{Name: "letStatement"}}
	statement.Node.AddChild(a.popToken())

	token := a.popToken()
	if err := assertToken(token, TokenTypeIdentifier); err != nil {
		return nil, err
	}
	statement.VarName = token.Value
	statement.Node.AddChild(token)

	var err error
	if checkToken(a.topToken(), TokenTypeSymbol, "[") {
		statement.Node.AddChild(a.popToken())

		statement.Index, err = a.parseExpression()
		if err != nil {
			return nil, err
		}
		statement.Node.AddChild(statement.Index)

		token := a.popToken()
		if err := assertToken(token, TokenTypeSymbol, "]"); err != nil {
			return nil, err
		}
		statement.Node.AddChild(token)
	}

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "="); err != nil {
		return nil, err
	}
	statement.Node.AddChild(token)

	val, err := a.parseExpression()
	if err != nil {
		return nil, err
	}
	statement.VarValue = *val
	statement.Node.AddChild(val)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, ";"); err != nil {
		return nil, err
	}
	statement.Node.AddChild(token)

	return &statement, nil
}

func (a *analyzer) parseIfStatement() (*IfStatement, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "if") {
		return nil, nil
	}
	statement := IfStatement{Node: Node{Name: "ifStatement"}}
	statement.Node.AddChild(a.popToken())

	token := a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "("); err != nil {
		return nil, err
	}
	statement.Node.AddChild(token)

	cond, err := a.parseExpression()
	if err != nil {
		return nil, err
	}
	statement.Condition = *cond
	statement.Node.AddChild(cond)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, ")"); err != nil {
		return nil, err
	}
	statement.Node.AddChild(token)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "{"); err != nil {
		return nil, err
	}
	statement.Node.AddChild(token)

	statements, err := a.parseStatements()
	if err != nil {
		return nil, err
	}
	statement.IfStatements = *statements
	statement.Node.AddChild(statements)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "}"); err != nil {
		return nil, err
	}
	statement.Node.AddChild(token)

	var elseStatements *Statements
	if checkToken(a.topToken(), TokenTypeKeyword, "else") {
		statement.Node.AddChild(a.popToken())

		token = a.popToken()
		if err := assertToken(token, TokenTypeSymbol, "{"); err != nil {
			return nil, err
		}
		statement.Node.AddChild(token)

		elseStatements, err = a.parseStatements()
		if err != nil {
			return nil, err
		}
		statement.ElseStatements = *elseStatements
		statement.Node.AddChild(elseStatements)

		token := a.popToken()
		if err := assertToken(token, TokenTypeSymbol, "}"); err != nil {
			return nil, err
		}
		statement.Node.AddChild(token)
	}

	return &statement, nil
}

func (a *analyzer) parseWhileStatement() (*WhileStatement, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "while") {
		return nil, nil
	}
	statement := WhileStatement{Node: Node{Name: "whileStatement"}}
	statement.Node.AddChild(a.popToken())

	token := a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "("); err != nil {
		return nil, err
	}
	statement.Node.AddChild(token)

	cond, err := a.parseExpression()
	if err != nil {
		return nil, err
	}
	statement.Condition = *cond
	statement.Node.AddChild(cond)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, ")"); err != nil {
		return nil, err
	}
	statement.Node.AddChild(token)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "{"); err != nil {
		return nil, err
	}
	statement.Node.AddChild(token)

	statements, err := a.parseStatements()
	if err != nil {
		return nil, err
	}
	statement.Statements = *statements
	statement.Node.AddChild(statements)

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "}"); err != nil {
		return nil, err
	}
	statement.Node.AddChild(token)

	return &statement, nil
}

func (a *analyzer) parseDoStatement() (*DoStatement, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "do") {
		return nil, nil
	}
	statement := DoStatement{Node: Node{Name: "doStatement"}}
	statement.Node.AddChild(a.popToken())

	call, err := a.parseSubroutineCall()
	if err != nil {
		return nil, err
	}
	statement.SubroutineCall = *call
	statement.Node.AddChild(call)

	token := a.popToken()
	if err := assertToken(token, TokenTypeSymbol, ";"); err != nil {
		return nil, err
	}
	statement.Node.AddChild(token)

	return &statement, nil
}

func (a *analyzer) parseReturnStatement() (*ReturnStatement, error) {
	if !checkToken(a.topToken(), TokenTypeKeyword, "return") {
		return nil, nil
	}
	statement := ReturnStatement{Node: Node{Name: "returnStatement"}}
	statement.Node.AddChild(a.popToken())

	if checkToken(a.topToken(), TokenTypeSymbol, ";") {
		statement.Node.AddChild(a.popToken())
		return &statement, nil
	}

	exp, err := a.parseExpression()
	if err != nil {
		return nil, err
	}
	statement.Expression = exp
	statement.Node.AddChild(exp)

	token := a.popToken()
	if err := assertToken(token, TokenTypeSymbol, ";"); err != nil {
		return nil, err
	}
	statement.Node.AddChild(token)

	return &statement, nil
}

func (a *analyzer) parseExpression() (*Expression, error) {

	exp := Expression{Node: Node{Name: "expression"}}

	term, err := a.parseTerm()
	if err != nil {
		return nil, err
	}
	exp.Term = *term
	exp.Node.AddChild(term)

	for {
		token := a.topToken()
		if checkToken(token, TokenTypeSymbol, "+", "-", "*", "/", "&", "|", "<", ">", "=") {
			op := Op(token.Value)
			exp.Node.AddChild(a.popToken())

			term, err := a.parseTerm()
			if err != nil {
				return nil, err
			}
			exp.Node.AddChild(term)

			exp.Tail = append(exp.Tail, ExpressionTail{Op: op, Term: *term})
			continue
		}
		break
	}

	return &exp, nil
}

func (a *analyzer) parseTerm() (*Term, error) {

	xml := Node{Name: "term"}

	token := a.popToken()
	switch {
	case checkToken(token, TokenTypeIntegerConst):
		i, err := strconv.ParseInt(token.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("integerConstant parse error: %+v", token)
		}
		xml.AddChild(token)
		return &Term{Type: TermTypeIntegerConst, IntegerConst: &i, Node: xml}, nil

	case checkToken(token, TokenTypeStringConst):
		xml.AddChild(token)
		return &Term{Type: TermTypeStringConst, StringConst: &token.Value, Node: xml}, nil

	case checkToken(token, TokenTypeKeyword, "true", "false", "null", "this"):
		xml.AddChild(token)
		return &Term{Type: TermTypeKeywordConst, KeywordConstant: &token.Value, Node: xml}, nil

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
			return &Term{Type: TermTypeSubroutineCall, SubroutineCall: call, Node: xml}, nil

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
			return &Term{Type: TermTypeVarNameIndex, VarName: &token.Value, Index: exp, Node: xml}, nil

		default:
			xml.AddChild(token)
			return &Term{Type: TermTypeVarName, VarName: &token.Value, Node: xml}, nil
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
		return &Term{Type: TermTypeExpression, Expression: exp, Node: xml}, nil

	case checkToken(token, TokenTypeSymbol, "-", "~"):
		xml.AddChild(token)

		op := UnaryOp(token.Value)
		term, err := a.parseTerm()
		if err != nil {
			return nil, err
		}
		xml.AddChild(term)
		return &Term{Type: TermTypeUnaryOp, UnaryOp: &op, UnaryOpTerm: term, Node: xml}, nil
	default:
		return nil, fmt.Errorf("term is expected: %+v", token)
	}
}

func (a *analyzer) parseSubroutineCall() (*SubroutineCall, error) {
	token := a.popToken()
	if err := assertToken(token, TokenTypeIdentifier); err != nil {
		return nil, err
	}

	call := SubroutineCall{Node: Node{SkipLayer: true}}
	call.Node.AddChild(token)

	next := a.topToken()
	if err := assertToken(next, TokenTypeSymbol, "(", "."); err != nil {
		return nil, err
	}

	switch next.Value {
	case "(":
		call.SubroutineName = token.Value
	case ".":
		call.Node.AddChild(a.popToken())
		call.ClassOrVarName = &token.Value

		token := a.popToken()
		if err := assertToken(token, TokenTypeIdentifier); err != nil {
			return nil, err
		}
		call.Node.AddChild(token)
	}

	token = a.popToken()
	if err := assertToken(token, TokenTypeSymbol, "("); err != nil {
		return nil, err
	}
	call.Node.AddChild(token)

	exps, err := a.parseExpressionList()
	if err != nil {
		return nil, err
	}
	call.ExpressionList = *exps
	call.Node.AddChild(exps)
	call.Node.AddChild(a.popToken()) // ")"

	return &call, nil
}

func (a *analyzer) parseExpressionList() (*ExpressionList, error) {
	exps := ExpressionList{Node: Node{Name: "expressionList", Children: []Node{}}}
	if checkToken(a.topToken(), TokenTypeSymbol, ")") {
		return &exps, nil
	}

	for {
		exp, err := a.parseExpression()
		if err != nil {
			return nil, err
		}
		exps.Node.AddChild(exp)

		exps.Expressions = append(exps.Expressions, *exp)

		token := a.topToken()
		if err := assertToken(token, TokenTypeSymbol, ",", ")"); err != nil {
			return nil, err
		}
		if token.Value == "," {
			exps.Node.AddChild(a.popToken())
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

func (x *Class) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *ClassVarDec) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *SubroutineDec) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *ParameterList) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *Parameter) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *SubroutineBody) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *VarDec) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *Statements) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *Statement) ToNode() *Node {
	if x == nil {
		return nil
	}
	switch x.Type {
	case StatementTypeLet:
		return x.LetStatement.ToNode()
	case StatementTypeIf:
		return x.IfStatement.ToNode()
	case StatementTypeWhile:
		return x.WhileStatement.ToNode()
	case StatementTypeDo:
		return x.DoStatement.ToNode()
	case StatementTypeReturn:
		return x.ReturnStatement.ToNode()
	default:
		return nil
	}
}
func (x *LetStatement) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *IfStatement) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *WhileStatement) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *DoStatement) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *ReturnStatement) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *Expression) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *Term) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *SubroutineCall) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}
func (x *ExpressionList) ToNode() *Node {
	if x == nil {
		return nil
	}
	return &x.Node
}

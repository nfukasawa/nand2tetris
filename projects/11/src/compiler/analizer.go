package compiler

import (
	"fmt"
	"strconv"
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
	dec.ParameterList = *params
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

	node := Node{Name: "term"}

	token := a.popToken()
	switch {
	case checkToken(token, TokenTypeIntegerConst):
		i, err := strconv.ParseInt(token.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("integerConstant parse error: %+v", token)
		}
		node.AddChild(token)
		return &Term{Type: TermTypeIntegerConst, IntegerConst: &i, Node: node}, nil

	case checkToken(token, TokenTypeStringConst):
		node.AddChild(token)
		return &Term{Type: TermTypeStringConst, StringConst: &token.Value, Node: node}, nil

	case checkToken(token, TokenTypeKeyword, "true", "false", "null", "this"):
		node.AddChild(token)
		return &Term{Type: TermTypeKeywordConst, KeywordConstant: &token.Value, Node: node}, nil

	case checkToken(token, TokenTypeIdentifier):
		next := a.topToken()
		switch {
		case checkToken(next, TokenTypeSymbol, "(", "."):
			a.pushToken(*token)
			call, err := a.parseSubroutineCall()
			if err != nil {
				return nil, err
			}
			node.AddChild(call)
			return &Term{Type: TermTypeSubroutineCall, SubroutineCall: call, Node: node}, nil

		case checkToken(next, TokenTypeSymbol, "["):
			node.AddChild(token)
			node.AddChild(a.popToken())
			exp, err := a.parseExpression()
			if err != nil {
				return nil, err
			}
			node.AddChild(exp)
			token := a.popToken()
			if err := assertToken(token, TokenTypeSymbol, "]"); err != nil {
				return nil, err
			}
			node.AddChild(token)
			return &Term{Type: TermTypeVarNameIndex, VarName: &token.Value, Index: exp, Node: node}, nil

		default:
			node.AddChild(token)
			return &Term{Type: TermTypeVarName, VarName: &token.Value, Node: node}, nil
		}

	case checkToken(token, TokenTypeSymbol, "("):
		node.AddChild(token)
		exp, err := a.parseExpression()
		if err != nil {
			return nil, err
		}
		node.AddChild(exp)
		token := a.popToken()
		if err := assertToken(token, TokenTypeSymbol, ")"); err != nil {
			return nil, err
		}
		node.AddChild(token)
		return &Term{Type: TermTypeExpression, Expression: exp, Node: node}, nil

	case checkToken(token, TokenTypeSymbol, "-", "~"):
		node.AddChild(token)

		op := UnaryOp(token.Value)
		term, err := a.parseTerm()
		if err != nil {
			return nil, err
		}
		node.AddChild(term)
		return &Term{Type: TermTypeUnaryOp, UnaryOp: &op, UnaryOpTerm: term, Node: node}, nil
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
		call.Receiver = &token.Value

		token := a.popToken()
		if err := assertToken(token, TokenTypeIdentifier); err != nil {
			return nil, err
		}
		call.SubroutineName = token.Value
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

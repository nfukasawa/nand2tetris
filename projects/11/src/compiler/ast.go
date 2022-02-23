package compiler

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
	Receiver       *string
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

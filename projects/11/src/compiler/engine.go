package compiler

import "fmt"

func Compile(vm *JackVM, cls *Class) error {
	newEngine(vm, cls).compile()
	return vm.Err()
}

type engine struct {
	vm                *JackVM
	class             *Class
	currentSubroutine SubRoutineType
	symbols           SymbolTable
}

func newEngine(vm *JackVM, class *Class) *engine {
	return &engine{
		vm:      vm,
		class:   class,
		symbols: NewSymbolTable(),
	}
}

func (e *engine) compile() {
	numField := e.defineClassVarSymbols(e.class)

	for _, dec := range e.class.SubRoutineDecs {
		numVar := e.defineSubroutineSymbols(e.class, &dec)

		e.vm.WriteFunction(e.class.ClassName+"."+dec.SubroutineName, numVar)

		e.currentSubroutine = dec.SubRoutineType
		switch dec.SubRoutineType {
		case SubRoutineTypeConstructor:
			e.vm.WritePush(VMSegCONST, numField)
			e.vm.WriteCall("Memory.alloc", 1)
			e.vm.WritePop(VMSegPOINTER, 0)

		case SubRoutineTypeMethod:
			e.vm.WritePush(VMSegARG, 0)
			e.vm.WritePop(VMSegPOINTER, 0)

		case SubRoutineTypeFunction:
		}

		lbl := newLabel(e.class.ClassName + "." + dec.SubroutineName)
		for _, s := range dec.SubroutineBody.Statements.Statements {
			e.compileStatement(&s, lbl)
		}
	}
}

func (e *engine) defineClassVarSymbols(cls *Class) (numField int64) {
	var numStatic int64
	for _, dec := range cls.ClassVarDecs {
		switch dec.ClassVarDecType {
		case ClassVarDecTypeField:
			for _, name := range dec.VarNames {
				e.symbols.Define(name, SymKind(dec.ClassVarDecType), string(dec.VarType), numField)
				numField++
			}
		case ClassVarDecTypeStatic:
			for _, name := range dec.VarNames {
				e.symbols.Define(name, SymKind(dec.ClassVarDecType), string(dec.VarType), numStatic)
				numStatic++
			}
		}
	}
	return numField
}

func (e *engine) defineSubroutineSymbols(cls *Class, dec *SubroutineDec) (numVar int64) {
	e.symbols.Subroutine()

	var numParam int64
	if dec.SubRoutineType == SubRoutineTypeMethod {
		numParam++
	}

	for _, param := range dec.ParameterList.Paramters {
		e.symbols.Define(param.VarName, SymKindArg, string(param.VarType), numParam)
		numParam++
	}

	for _, dec := range dec.SubroutineBody.VarDecs {
		for _, name := range dec.VarNames {
			e.symbols.Define(name, SymKindVar, string(dec.VarType), numVar)
			numVar++
		}
	}
	return numVar
}

func (e *engine) compileStatement(s *Statement, label *label) {
	switch s.Type {
	case StatementTypeLet:
		e.compileLetStatement(s.LetStatement)
	case StatementTypeIf:
		e.compileIfStatement(s.IfStatement, label)
	case StatementTypeWhile:
		e.compileWhileStatement(s.WhileStatement, label)
	case StatementTypeDo:
		e.compileDoStatement(s.DoStatement)
	case StatementTypeReturn:
		e.compileReturnStatement(s.ReturnStatement)
	}
}

func (e *engine) compileLetStatement(s *LetStatement) {
	e.compileExpression(&s.VarValue)

	if s.Index != nil {
		e.vm.WritePush(sym2VM(e.symbols.Get(s.VarName)))
		e.compileExpression(s.Index)
		e.vm.WriteArithmetic(VMCmdADD)
		e.vm.WritePop(VMSegPOINTER, 1)
		e.vm.WritePop(VMSegTHAT, 0)
		return
	}

	e.vm.WritePop(sym2VM(e.symbols.Get(s.VarName)))
}

func (e *engine) compileIfStatement(s *IfStatement, label *label) {
	elseL := label.Get()
	endL := label.Get()

	e.compileExpression(&s.Condition)
	e.vm.WriteArithmetic(VMCmdNOT)
	e.vm.WriteIfGoto(elseL)
	for _, s := range s.IfStatements.Statements {
		e.compileStatement(&s, label)
	}
	e.vm.WriteGoto(endL)

	e.vm.WriteLabel(elseL)
	for _, s := range s.ElseStatements.Statements {
		e.compileStatement(&s, label)
	}
	e.vm.WriteLabel(endL)
}

func (e *engine) compileWhileStatement(s *WhileStatement, label *label) {
	loopL := label.Get()
	endL := label.Get()

	e.vm.WriteLabel(loopL)
	e.compileExpression(&s.Condition)
	e.vm.WriteArithmetic(VMCmdNOT)
	e.vm.WriteIfGoto(endL)
	for _, s := range s.Statements.Statements {
		e.compileStatement(&s, label)
	}
	e.vm.WriteGoto(loopL)
	e.vm.WriteLabel(endL)
}

func (e *engine) compileDoStatement(s *DoStatement) {
	e.compileSubroutineCall(&s.SubroutineCall)
	e.vm.WritePop(VMSegTEMP, 0)
}

func (e *engine) compileReturnStatement(s *ReturnStatement) {
	if s.Expression != nil {
		e.compileExpression(s.Expression)
	} else {
		e.vm.WritePush(VMSegCONST, 0)
	}
	e.vm.WriteReturn()
}

func (e *engine) compileExpression(exp *Expression) {
	e.compileTerm(&exp.Term)
	for _, t := range exp.Tail {
		e.compileTerm(&t.Term)
		e.compileOp(&t.Op)
	}
}

func (e *engine) compileTerm(t *Term) {
	switch t.Type {
	case TermTypeIntegerConst:
		e.vm.WritePush(VMSegCONST, *t.IntegerConst)

	case TermTypeStringConst:
		e.vm.WritePush(VMSegCONST, int64(len(*t.StringConst)))
		e.vm.WriteCall("String.new", 1)
		for _, c := range *t.StringConst {
			e.vm.WritePush(VMSegCONST, int64(c))
			e.vm.WriteCall("String.appendChar", 2)
		}

	case TermTypeKeywordConst:
		switch *t.KeywordConstant {
		case "true":
			e.vm.WritePush(VMSegCONST, 1)
			e.vm.WriteArithmetic(VMCmdNEG)
		case "false", "null":
			e.vm.WritePush(VMSegCONST, 0)
		case "this":
			e.vm.WritePush(VMSegPOINTER, 0)
		}

	case TermTypeVarName:
		e.vm.WritePush(sym2VM(e.symbols.Get(*t.VarName)))

	case TermTypeVarNameIndex:
		e.vm.WritePush(sym2VM(e.symbols.Get(*t.VarName)))
		e.compileExpression(t.Index)
		e.vm.WriteArithmetic(VMCmdADD)
		e.vm.WritePop(VMSegPOINTER, 1)
		e.vm.WritePush(VMSegTHAT, 0)

	case TermTypeSubroutineCall:
		e.compileSubroutineCall(t.SubroutineCall)

	case TermTypeExpression:
		e.compileExpression(t.Expression)

	case TermTypeUnaryOp:
		e.compileTerm(t.UnaryOpTerm)
		e.compileUnaryOp(t.UnaryOp)
	}
}

func (e *engine) compileSubroutineCall(call *SubroutineCall) {
	var name string
	numArgs := int64(len(call.ExpressionList.Expressions))

	if call.Receiver == nil {
		// クラス内のメソッド呼び出し
		// -> methodName()
		switch e.currentSubroutine {
		case SubRoutineTypeConstructor:
			e.vm.WritePush(VMSegPOINTER, 0)
		case SubRoutineTypeMethod:
			e.vm.WritePush(VMSegARG, 0)
		}
		name = e.class.ClassName + "." + call.SubroutineName
		numArgs++
	} else {
		sym := e.symbols.Get(*call.Receiver)
		if sym != nil {
			// 他のクラスのメソッド呼び出し
			// -> varName.methodName()
			e.vm.WritePush(sym2VM(sym))
			name = sym.Type + "." + call.SubroutineName
			numArgs++
		} else {
			// コンストラクタ/ファンクション呼び出し
			// -> className.subroutineName()
			name = *call.Receiver + "." + call.SubroutineName
			ClassMethodCalled(*call.Receiver)
		}
	}

	for _, exp := range call.ExpressionList.Expressions {
		e.compileExpression(&exp)
	}
	e.vm.WriteCall(name, numArgs)
}

func (e *engine) compileOp(op *Op) {
	switch *op {
	case "+":
		e.vm.WriteArithmetic(VMCmdADD)
	case "-":
		e.vm.WriteArithmetic(VMCmdSUB)
	case "*":
		e.vm.WriteCall("Math.multiply", 2)
	case "/":
		e.vm.WriteCall("Math.divide", 2)
	case "&":
		e.vm.WriteArithmetic(VMCmdAND)
	case "|":
		e.vm.WriteArithmetic(VMCmdOR)
	case "<":
		e.vm.WriteArithmetic(VMCmdLT)
	case ">":
		e.vm.WriteArithmetic(VMCmdGT)
	case "=":
		e.vm.WriteArithmetic(VMCmdEQ)
	}
}

func (e *engine) compileUnaryOp(op *UnaryOp) {
	switch *op {
	case "-":
		e.vm.WriteArithmetic(VMCmdNEG)
	case "~":
		e.vm.WriteArithmetic(VMCmdNOT)
	}
}

func sym2VM(s *Symbol) (VMSeg, int64) {
	switch s.Kind {
	case SymKindStatic:
		return VMSegSTATIC, s.Index
	case SymKindField:
		return VMSegTHIS, s.Index
	case SymKindArg:
		return VMSegARG, s.Index
	case SymKindVar:
		return VMSegLOCAL, s.Index
	default:
		return VMSeg(""), -1
	}
}

type label struct {
	base  string
	index int
}

func newLabel(base string) *label {
	return &label{base: base}
}

func (l *label) Get() string {
	ret := fmt.Sprintf("%s.%d", l.base, l.index)
	l.index++
	return ret
}

package compiler

type Symbol struct {
	Kind  SymKind
	Type  string
	Index int64
}

type SymKind string

const (
	SymKindStatic SymKind = "static"
	SymKindField  SymKind = "field"
	SymKindArg    SymKind = "argument"
	SymKindVar    SymKind = "var"
	SymKindNone   SymKind = ""
)

type SymbolTable struct {
	classTable      map[string]Symbol
	subroutineTable map[string]Symbol
}

func NewSymbolTable() SymbolTable {
	return SymbolTable{
		classTable: map[string]Symbol{},
	}
}

func (t *SymbolTable) Subroutine() {
	t.subroutineTable = map[string]Symbol{}
}

func (t *SymbolTable) Define(name string, kind SymKind, ty string, index int64) {
	sym := Symbol{Kind: kind, Type: ty, Index: index}
	if t.subroutineTable != nil {
		t.subroutineTable[name] = sym
	}
	t.classTable[name] = sym
}

func (t *SymbolTable) Get(name string) *Symbol {
	v, ok := t.subroutineTable[name]
	if ok {
		return &v
	}
	v, ok = t.classTable[name]
	if ok {
		return &v
	}
	return nil
}

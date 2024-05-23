package compiler

/*
建立符号表,将全局或局部标识符与特定数字相关联,
并获取已给定标识符相关联的数字
*/
type SymbolScope string // 符号作用域

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
)

type Symbol struct {
	Name  string // 符号的名称。
	Scope SymbolScope
	Index int // 表示符号在特定作用域中的索引
}

type SymbolTable struct { // 表示符号表
	Outer *SymbolTable

	store          map[string]Symbol // 将符号的名称（字符串）映射到 Symbol 结构体。
	numDefinitions int               // 是一个计数器，跟踪定义的符号数量。
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func (s *SymbolTable) Define(name string) Symbol { // 将标识符作为参数,创建定义并返回Symbol
	symbol := Symbol{Name: name, Index: s.numDefinitions}
	if s.Outer == nil { // Outer为空则设置为全局变量
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) { // 将一个先前定义的标识符交给符号表,并返回关联的symbol
	obj, ok := s.store[name]
	if !ok && s.Outer != nil {
		obj, ok = s.Outer.Resolve(name)
		return obj, ok
	}
	return obj, ok
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol { // 定义内置函数的作用域
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	s.store[name] = symbol
	return symbol
}

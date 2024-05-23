package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"

	"my.com/myfile/ast"
	"my.com/myfile/code"
)

type ObjectType string //增加了代码的可读性

const (
	NULL_OBJ  = "NULL"
	ERROR_OBJ = "ERROR"

	INTEGER_OBJ = "INTEGER"
	STRING_OBJ  = "STRING"
	BOOLEAN_OBJ = "BOOLEAN"

	RETURN_VALUE_OBJ   = "RETURN_VALUE"
	CONTINUE_VALUE_OBJ = "CONTINUE_VALUE"
	BREAK_VALUE_OBJ    = "BREAK_VALUE"
	BUILTIN_OBJ        = "BUILTIN"

	FUNCTION_OBJ = "FUNCTION" // 保存函数的ast节点
	ARRAY_OBJ    = "ARRAY"
	HASH_OBJ     = "HASH"

	COMPILED_FUNCTION_OBJ = "COMPILED_FUNCTION_OBJ" // 保存函数的字节码
)

// Object 定义了Object接口，接口提供了Type方法和Inspect方法
type Object interface {
	Type() ObjectType
	Inspect() string
	ToBoolean() bool
}

// Integer 整数的处理方法
type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) } //fmt.Sprintf 函数是一种通用的函数，用于将格式化的字符串生成并返回，而不是直接打印到标准输出。
func (i *Integer) ToBoolean() bool {
	if i.Value != 0 {
		return true
	} else {
		return false
	}
}

// Boolean 这个函数非常灵活，支持多种格式化的占位符，用于将不同类型的值转化为字符串。
// 布尔值的处理方法
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) ToBoolean() bool  { return b.Value }

// Null 空的处理方法
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }
func (n *Null) ToBoolean() bool  { return false }

// ReturnValue 返回类型的处理方法
type ReturnValue struct { //返回其值
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
func (rv *ReturnValue) ToBoolean() bool  { return rv.Value.ToBoolean() }

// Error 错误类型
type Error struct {
	Message string
}

// BreakValue break的处理方法
type BreakValue struct {
	Value Object
}

func (bv *BreakValue) Type() ObjectType { return BREAK_VALUE_OBJ }
func (bv *BreakValue) Inspect() string  { return bv.Value.Inspect() }
func (bv *BreakValue) ToBoolean() bool  { return bv.Value.ToBoolean() }

// ContinueValue continue的处理方法
type ContinueValue struct {
	Value Object
}

func (cv *ContinueValue) Type() ObjectType { return CONTINUE_VALUE_OBJ }
func (cv *ContinueValue) Inspect() string  { return cv.Value.Inspect() }
func (cv *ContinueValue) ToBoolean() bool  { return cv.Value.ToBoolean() }

// Type 错误类型的处理方法
func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
func (e *Error) ToBoolean() bool  { return false }

// Function 函数的处理方法
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}
func (f *Function) ToBoolean() bool { return true }

// String 字符串的处理方法
type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) ToBoolean() bool {
	if s.Value != "" {
		return true
	} else {
		return false
	}
}

// BuiltinFunction 接收任意数量的参数
type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
func (b *Builtin) ToBoolean() bool  { return true }

type Array struct {
	Elements []Object
}

func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}
func (ao *Array) ToBoolean() bool { return true }

// 哈希表实现
type HashKey struct {
	Type  ObjectType
	Value uint64
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Name  string
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
func (h *Hash) ToBoolean() bool { return true }
func (h *Hash) HashKey() HashKey {
	return HashKey{}
}

// hashable 接口实现
type Hashable interface {
	HashKey() HashKey
}

func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

type CompiledFunction struct {
	Instructions  code.Instructions
	NumLocals     int // 反馈函数有多少个局部绑定
	NumParameters int
}

func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION_OBJ }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}
func (cf *CompiledFunction) ToBoolean() bool { return true }

package compiler

import (
	"fmt"
	"my.com/myfile/ast"
	"my.com/myfile/code"
	"my.com/myfile/object"
	"sort"
)

type Compiler struct {
	//instructions        code.Instructions  // 指令
	constants []object.Object // 常量池
	//lastInstruction     EmittedInstruction // 发出的最后一条指令
	//previousInstruction EmittedInstruction // 发出的倒数第二条指令
	SymbolTable *SymbolTable

	scopes     []CompilationScope // 存放函数作用域
	scopeIndex int
}

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

func New() *Compiler { // 返回Compiler结构指针
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	symbolTable := NewSymbolTable()

	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	return &Compiler{
		constants:   []object.Object{},
		SymbolTable: NewSymbolTable(),
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
	}
}

func (c *Compiler) Compile(node ast.Node) error { // 编译器
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement: // 表达式
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop) // 每次执行表达式后执行一次弹栈操作清理栈
	case *ast.InfixExpression: // 中缀表达式
		if node.Operator == "<" { // 实现<的逆操作，也就是>
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
		}
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpGreaterThan)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IntegerLiteral: // 整数字面量,利用object中已有的对象简化工作
		integer := &object.Integer{Value: node.Value}   // 求值
		c.emit(code.OpConstant, c.addConstant(integer)) // 生成opConstant指令
	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.PrefixExpression: // 前缀表达式
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.IfExpression: // 处理if表达式
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)
		err = c.Compile(node.Consequence)
		if err != nil {
			return nil
		}

		if c.lastInstructionIs(code.OpPop) { // if语句块的结尾不能执行出栈执行，栈里必须有一个结果，因此这里需要删除最后一个的OpPop指令
			c.removeLastPop()
		}
		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos) // 将位置在jumpNotTruthyPos处的OpJumpNotTruthy指令的操作数替换为afterConsequencePos

		if node.Alternative == nil { // 如果else语句为空,就压栈Null,避免if语句不成立出栈的时候栈没有内容的情况
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)
	case *ast.BlockStatement: // 处理block语句块
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		symbol := c.SymbolTable.Define(node.Name.Value)
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}
	case *ast.Identifier: // 解析变量名
		symbol, ok := c.SymbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		c.loadSymbol(symbol)
	case *ast.ArrayLiteral: // 数组
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return nil
			}
		}

		c.emit(code.OpArray, len(node.Elements))
	case *ast.HashLiteral: // 哈希
		keys := []ast.Expression{}
		for k := range node.Pairs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { // 对键值对排序,避免顺序不一致
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}
			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}

		c.emit(code.OpHash, len(node.Pairs)*2) // 操作数为键和值的数量
	case *ast.IndexExpression: // 索引字面量
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)
	case *ast.FunctionLiteral: // 函数字面量,在编译函数时更改发出指令的存储位置
		c.enterScope()

		for _, p := range node.Parameters {
			c.SymbolTable.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) { // 函数最后一条的出栈指令用return代替
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) { // 考虑到函数没有任何语句的情况
			c.emit(code.OpReturn)
		}

		numLocals := c.SymbolTable.numDefinitions // 计数局部变量
		instructions := c.leaveScope()

		compiledFn := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}
		c.emit(code.OpConstant, c.addConstant(compiledFn))

	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)

	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, a := range node.Arguments {
			err := c.Compile(a)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))
	}
	return nil
}

type Bytecode struct {
	Instructions code.Instructions // 字节码
	Constants    []object.Object   // 切片类型，常量池
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.Object) int { // 将求值结果添加到常量池中
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1 // 添加到编译器constants切片末尾，返回其在constants切片中的索引来为其提供标识符，用作opConstant指令的操作数
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int { // 根据操作码和操作数生成对应的字节码序列
	// 生成指令
	ins := code.Make(op, operands...)

	// 将指令添加到指令集中
	pos := c.addInstruction(ins)

	// 记录最后生成的指令和位置
	c.setLastInstruction(op, pos)

	// 返回新指令的位置
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	// 获取新指令的起始位置
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)
	// 将新指令添加到指令集中
	c.scopes[c.scopeIndex].instructions = updatedInstructions

	// 返回新指令的起始位置
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) { // 获得最后发出指令的两个操作码
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool { // 判断上一个指令是否是出栈指令
	if len(c.currentInstructions()) == 0 {
		return false
	}

	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() { // 移除最后一条指令，做法是将存放指令的栈内容后移
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	new := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) { // 替换指令
	ins := c.currentInstructions()

	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) { // 创建新指令，并调用replaceInstruction替换指令
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler { // 新的编译器初始化函数,
	compiler := New()
	compiler.SymbolTable = s
	compiler.constants = constants
	return compiler
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

func (c *Compiler) currentInstructions() code.Instructions { // 返回当前作用域
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) enterScope() { // 添加函数作用域
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++

	c.SymbolTable = NewEnclosedSymbolTable(c.SymbolTable)
}

func (c *Compiler) leaveScope() code.Instructions { // 删除函数作用域
	instructions := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.SymbolTable = c.SymbolTable.Outer

	return instructions
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))

	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

func (c *Compiler) loadSymbol(s Symbol) { // 判断符号表表达内容
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, s.Index)
	}
}

package vm

// 定义虚拟机

import (
	"fmt"
	"my.com/myfile/code"
	"my.com/myfile/compiler"
	"my.com/myfile/object"
)

const StackSize = 2048
const GlobalsSize = 65536 // 定义全局变量的数量绑定上线
const MaxFrames = 1024

// 定义全局变量True,False,Null
var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type VM struct {
	constants []object.Object // 常量池
	//instructions code.Instructions // 指令

	stack   []object.Object // 栈
	sp      int             // 指向栈里下一个空闲槽，栈顶的值是stack[sp-1]
	globals []object.Object // 储存全局变量

	frames      []*Frame
	framesIndex int
}

func New(bytecode *compiler.Bytecode) *VM { // 创建栈
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		//instructions: bytecode.Instructions,
		constants: bytecode.Constants,

		stack:   make([]object.Object, StackSize),
		sp:      0,
		globals: make([]object.Object, GlobalsSize),

		frames:      frames,
		framesIndex: 1,
	}
}

func (vm *VM) StackTop() object.Object { // 访问栈顶元素
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error { // 运行虚拟机，执行操作并对每个操作结果进行压栈
	var ip int
	var ins code.Instructions
	//var op code.Opcode

	//for ip := 0; ip < len(vm.instructions); ip++ { // 取出指令
	//	op := code.Opcode(vm.instructions[ip]) // 也可以使用code.Lookup代替，但是速度慢
	//
	//
	//	}
	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op := code.Opcode(ins[ip])

		switch op {
		case code.OpConstant: // 从虚拟机的常量池中取出一个常量值，并将其压入到虚拟机的栈中
			constIndex := code.ReadUint16(ins[ip+1:]) // 这里也可以用code.ReadOperands代替ReadUint16，但是速度慢
			vm.currentFrame().ip += 2                 // opConstant宽度为2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv: // 加法
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpTrue: // 将bool值压栈
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan: // 比较运算，压栈
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := vm.executeBangOperator() // 执行布尔取反,压栈
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:])) // 解码操作数
			vm.currentFrame().ip = pos - 1          // 将指令指针ip设置为跳转指令的目标处
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2 // 跳过操作数的两字节

			condition := vm.pop()
			if !isTruthy(condition) { // 如果条件不为真，就需要执行跳转
				vm.currentFrame().ip = pos - 1
			}
		case code.OpNull:
			err := vm.push(Null)

			if err != nil {
				return err
			}
		case code.OpSetGlobal: // 设置全局变量
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal: // 获取全局变量
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpArray: // 数组压栈
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements

			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		case code.OpCall: // 调用函数
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			err := vm.executeCall(int(numArgs))
			if err != nil {
				return err
			}
		case code.OpReturnValue:
			returnValue := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetLocal: // 局部变量赋值
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			vm.stack[frame.basePointer+int(localIndex)] = vm.pop()
		case code.OpGetLocal: // 获取局部变量
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			err := vm.push(vm.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			builtinIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			definition := object.Builtins[builtinIndex]

			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isTruthy(obj object.Object) bool { // 判断是否为真
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null: // 当if嵌套并且内部产生null的时候,我们需要一个false
		return false

	default:
		return true
	}
}

func (vm *VM) push(o object.Object) error { // 压栈操作
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overFlow")
	}

	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object { // 出栈操作
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) LastPoppedStackElem() object.Object { // 返回最近弹栈的元素
	return vm.stack[vm.sp]
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error { // 做类型断言
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()
	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}

	return fmt.Errorf("unsupported types for binary operator: %s %s", leftType, rightType)
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error { // 处理整数操作
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unkoown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeComparison(op code.Opcode) error { // 执行比较
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ { // 只有当左右字面量都为整型时才进行比较
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBooleanToBooleanObject(right == left)) // 转换go的布尔类型
	case code.OpNotEqual:
		return vm.push(nativeBooleanToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left object.Object, right object.Object) error { // 整数比较运算
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBooleanToBooleanObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBooleanToBooleanObject(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBooleanToBooleanObject(rightValue < leftValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeBangOperator() error { // 布尔取反
	operand := vm.pop() // 操作数出栈

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()
	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for nagation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func nativeBooleanToBooleanObject(input bool) object.Object { // 转换bool为Wizard的bool类型
	if input {
		return True
	}
	return False
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM { // 新的虚拟机初始化函数
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) executeBinaryStringOperation( // 字符串拼接运算
	op code.Opcode,
	left, right object.Object,
) error {
	if op != code.OpAdd {
		return fmt.Errorf("enknown string operator: %d", op)
	}

	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	return vm.push(&object.String{Value: leftValue + rightValue})
}

func (vm *VM) buildArray(startIndex, endIndex int) object.Object { // 创建数组
	elements := make([]object.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	hashedParis := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusuable as hash key %s:", key.Type())
		}

		hashedParis[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedParis}, nil
}

func (vm *VM) executeIndexExpression(left object.Object, index object.Object) error { // 返回left索引为index的元素,对数组和哈希表单独处理
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array object.Object, index object.Object) error {
	arrayObject := array.(*object.Array) // go的类型断言,将array转换成*object.Array类型,若转换失败会引发异常
	i := index.(*object.Integer).Value

	max := int64(len(arrayObject.Elements) - 1)

	if i < 0 || i > max { // 如果索引位置不对,那么就把Null压栈
		return vm.push(Null)
	}

	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(hash object.Object, index object.Object) error {
	hashObject := hash.(*object.Hash) // go的类型断言,将array转换成*object.Hash类型,若转换失败会引发异常

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusble as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) currentFrame() *Frame { // 返回当前栈帧栈顶元素
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) popFrame() *Frame { // 从栈帧中出栈
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) pushFrame(f *Frame) { // 栈帧压栈
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) callFunction(fn *object.CompiledFunction, numArgs int) error { // 调用编译后的函数
	if numArgs != fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			fn.NumParameters, numArgs)
	}

	frame := NewFrame(fn, vm.sp-numArgs)
	vm.pushFrame(frame)

	vm.sp = frame.basePointer + fn.NumLocals

	return nil
}

func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error { // 调用内置函数
	args := vm.stack[vm.sp-numArgs : vm.sp]

	result := builtin.Fn(args...)
	vm.sp = vm.sp - numArgs - 1

	if result != nil {
		vm.push(result)
	} else {
		vm.push(Null)
	}

	return nil
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.CompiledFunction:
		return vm.callFunction(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function and non-built-in")
	}
}

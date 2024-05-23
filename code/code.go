package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte // 字节集合

type Opcode byte // 操作码

const (
	OpConstant    Opcode = iota // 以操作数为索引检索常量并压栈
	OpAdd                       // +
	OpPop                       // 出栈
	OpSub                       // -
	OpMul                       // *
	OpDiv                       // /
	OpTrue                      // true
	OpFalse                     // false
	OpEqual                     // ==
	OpNotEqual                  // !=
	OpGreaterThan               // >,因为是>的逆操作所以这里没有定义<
	// TODO: 实现<=,>=
	OpMinus         // 对整数取负
	OpBang          // 布尔值取反
	OpJumpNotTruthy // 不会真跳转
	OpJump          // 直接跳转
	OpNull          // Null,用于产生空值的情况
	OpGetGlobal     // 获取全局变量
	OpSetGlobal     // 定义全局变量
	OpGetLocal      // 获取局部变量
	OpSetLocal      // 定义局部变量
	OpArray
	OpHash
	OpIndex // 数组和哈希索引
	OpCall
	OpReturnValue
	OpReturn
	OpGetBuiltin
)

type Definition struct {
	Name          string // 操作数名车给
	OperandWidths []int  // 字节宽度
}

var definitions = map[Opcode]*Definition{ // 不用操作数宽度为0,反之为2
	OpConstant:      {"OpConstant", []int{2}},      // OpConstant占两字节
	OpAdd:           {"opAdd", []int{}},            // 空的整数切片，不需要操作数
	OpPop:           {"OpPop", []int{}},            // 弹栈，用于清理栈，每个表达式语句执行后都要执行这个操作码
	OpSub:           {"OpSub", []int{}},            // 减法操作
	OpMul:           {"OpMul", []int{}},            // 乘法
	OpDiv:           {"OpDiv", []int{}},            // 除法
	OpTrue:          {"OpTrue", []int{}},           // 布尔真
	OpFalse:         {"OpFalse", []int{}},          // 布尔假
	OpEqual:         {"OpEqual", []int{}},          // ==
	OpNotEqual:      {"OpNotEqual", []int{}},       // !=
	OpGreaterThan:   {"OpGreaterThan", []int{}},    // >
	OpMinus:         {"OpMinus", []int{}},          // -,直接操作栈顶元素，不用操作数
	OpBang:          {"OpBang", []int{}},           // !
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}}, // 两字节
	OpJump:          {"Opjump", []int{2}},          // 两字节
	OpNull:          {"OpNull", []int{}},
	OpGetGlobal:     {"OpGetGlobal", []int{2}},
	OpSetGlobal:     {"OpSetGlobal", []int{2}},
	OpArray:         {"OpArray", []int{2}}, // OpArray有一个操作数,即数组中的元素个数
	OpHash:          {"OpHash", []int{2}},
	OpIndex:         {"OpIndex", []int{}},       // 没有操作数
	OpCall:          {"OpCall", []int{1}},       // 运行位于栈顶的*object.CompiledFunction
	OpReturnValue:   {"OpReturnValue", []int{}}, // 函数return语句,没有操作数,返回栈顶元素
	OpReturn:        {"OpReturn", []int{}},      // 返回调用函数之前的逻辑
	OpSetLocal:      {"OpSetLocal", []int{1}},
	OpGetLocal:      {"OpGetLocal", []int{1}},
	OpGetBuiltin:    {"OpGetBuiltin", []int{1}},
}

func Lookup(op byte) (*Definition, error) { // 查找操作码
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

func Make(op Opcode, operands ...int) []byte { // 创建包含操作码和可选操作数的指令，...代表可选
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1

	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2: // 双字节
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o)) // 将一个 uint16 类型的整数值（o）以大端字节序的形式写入到一个字节切片（instruction）的指定位置（从 offset 开始）
		case 1: // 单字节
			instruction[offset] = byte(o)
		}
		offset += width
	}

	return instruction
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) { // Make的逆向操作，返回解码后的操作数
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}

		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 { // 从给定的 Instructions 类型变量中读取一个 16 位无符号整数，并使用大端序解释数据
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}
func (ins Instructions) String() string { // 返回指令的可读形式
	var out bytes.Buffer

	i := 0
	for i <= len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string { // 打印操作数
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match definition %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0: // 如果操作数数量为 0，直接返回指令的名称。
		return def.Name
	case 1: //  如果操作数数量为 1，返回指令名称和第一个操作数的字符串表示形式。
		return fmt.Sprintf("%s, %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

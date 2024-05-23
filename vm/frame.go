package vm

// 构成函数栈帧,独立于前面定义的栈

import (
	"my.com/myfile/code"
	"my.com/myfile/object"
)

type Frame struct { // 帧
	fn          *object.CompiledFunction // 指向帧已引用的已编译函数
	ip          int                      // 该帧指令指针
	basePointer int
}

func NewFrame(fn *object.CompiledFunction, basePointer int) *Frame {
	f := &Frame{
		fn:          fn,
		ip:          -1,
		basePointer: basePointer, // 基指针,指向栈帧底部
	}

	return f
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}

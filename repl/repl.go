package repl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"my.com/myfile/code"
	"my.com/myfile/compiler"
	"my.com/myfile/lexer"
	"my.com/myfile/object"
	"my.com/myfile/parser"
	"my.com/myfile/vm"
)

const PROMPT = ">> "
const MULTILINE_PROMPT = "... "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	var input strings.Builder

	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	fmt.Fprintf(out, PROMPT)
	for scanner.Scan() {
		line := scanner.Text()
		input.WriteString(line + "\n")

		if isCompleteInput(input.String()) {
			processInput(input.String(), out, &constants, globals, symbolTable)
			input.Reset()
			fmt.Fprintf(out, PROMPT)
		} else {
			fmt.Fprintf(out, MULTILINE_PROMPT)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(out, "error: %v\n", err)
	}
}
func printParserErrors(out io.Writer, errors []string) { //错误输出
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}

func IsComplete(input string) bool { //判断完整性，但是你在程序中间不能空两行，因为那样会被视为结束
	// 将输入字符串拆分成行
	lines := strings.Split(input, "\n")

	// 获取最后两行
	lastLine := lines[len(lines)-1]
	secondLastLine := lines[len(lines)-2]

	// 如果最后两行都是空字符串，则认为是完整的语句
	return strings.TrimSpace(lastLine) == "" && strings.TrimSpace(secondLastLine) == ""
}

func isCompleteInput(input string) bool {
	openBraces := 0
	openParens := 0

	for _, char := range input {
		switch char {
		case '{':
			openBraces++
		case '}':
			openBraces--
		case '(':
			openParens++
		case ')':
			openParens--
		}
	}

	return openBraces == 0 && openParens == 0
}

func processInput(input string, out io.Writer, constants *[]object.Object, globals []object.Object, symbolTable *compiler.SymbolTable) {
	if isDisassembleCommand(input) {
		handleDisassembleCommand(input, out)
	} else {
		handleNormalCommand(input, out, constants, globals, symbolTable)
	}
}

func isDisassembleCommand(input string) bool {
	return strings.HasPrefix(strings.TrimSpace(input), "dis(") && strings.HasSuffix(strings.TrimSpace(input), ")")
}

func handleNormalCommand(input string, out io.Writer, constants *[]object.Object, globals []object.Object, symbolTable *compiler.SymbolTable) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParserErrors(out, p.Errors())
		return
	}

	//comp := compiler.New()
	comp := compiler.NewWithState(symbolTable, *constants)
	err := comp.Compile(program)
	if err != nil {
		fmt.Fprintf(out, "编译失败:\n %s\n", err)
		return
	}

	code := comp.Bytecode()
	*constants = code.Constants
	//machine := vm.New(comp.Bytecode())
	machine := vm.NewWithGlobalsStore(code, globals)
	err = machine.Run()
	if err != nil {
		fmt.Fprintf(out, "执行失败\n %s\n", err)
		return
	}

	lastPopped := machine.LastPoppedStackElem()
	if lastPopped != nil {
		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")
	}
}

func handleDisassembleCommand(input string, out io.Writer) {
	codeStr := extractCodeFromDisCommand(input)
	if codeStr == "" {
		fmt.Fprintf(out, "请输入有效的代码以进行反汇编\n")
		return
	}

	l := lexer.New(codeStr)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParserErrors(out, p.Errors())
		return
	}

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		fmt.Fprintf(out, "编译失败:\n %s\n", err)
		return
	}

	bytecode := comp.Bytecode()
	disassembled := disassemble(bytecode)
	fmt.Fprintf(out, "字节码反汇编:\n%s\n", disassembled)
}

func disassemble(bytecode *compiler.Bytecode) string {
	var out strings.Builder

	out.WriteString("Instructions:\n")
	out.WriteString(formatInstructions(bytecode.Instructions))
	out.WriteString("\nConstants:\n")

	for i, constant := range bytecode.Constants {
		out.WriteString(fmt.Sprintf("%04d %s\n", i, constant.Inspect()))
	}

	return out.String()
}

func extractCodeFromDisCommand(input string) string {
	return strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(input), "dis("), ")")
}

func formatInstruction(def *code.Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match definition %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

func formatInstructions(ins code.Instructions) string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := code.Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := code.ReadOperands(def, ins[i+1:])
		if read > len(ins[i+1:]) {
			fmt.Fprintf(&out, "ERROR: unexpected end of instructions\n")
			break
		}

		fmt.Fprintf(&out, "%04d %s\n", i, formatInstruction(def, operands))
		i += 1 + read
	}

	return out.String()
}

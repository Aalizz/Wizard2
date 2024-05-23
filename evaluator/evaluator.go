package evaluator

import (
	"fmt"

	"my.com/myfile/ast"
	"my.com/myfile/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object { //repl调用的函数
	switch node := node.(type) { //观察这个switch语句，尽管对于不同的ast结构体有不同的处理函数，但是实际上都会返回一个接口

	// Statements
	case *ast.Program: //顶层节点，evalProgram遍历程序中的所有语句
		return evalProgram(node, env)

	case *ast.BlockStatement: //块语句
		return evalBlockStatement(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.BreakStatement:
		return &object.BreakValue{Value: NULL}

	case *ast.ContinueStatement:
		return &object.ContinueValue{Value: NULL}

	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		if val.Type() == object.HASH_OBJ {
			hashObj := val.(*object.Hash)
			hashObj.Name = node.Name.Value
			env.Set(hashObj.Name, hashObj)
		} else {
			env.Set(node.Name.Value, val)
		}

	// 表达式
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalInfixExpression(node.Operator, left, right)

	// 控制语句
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ForExpression:
		return evalForExpression(node, env)
	case *ast.WhileExpression:
		return evalWhileExpression(node, env)
	case *ast.Identifier:
		result := evalIdentifier(node, env)
		if result.Type() == object.HASH_OBJ {
			hashObj, ok := result.(*object.Hash)
			if !ok {
				return newError("identifier not found: " + node.Value)
			}
			env.Set(hashObj.Name, hashObj)
			return hashObj
		} else {
			return result
		}
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}

		// 表达式处理
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env) //
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)

		// 字符串求值
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

		// 数组表达式求值
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	// 处理下标读取
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}

	return nil
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements { //通过循环遍历所有语句，对每个语句调用Eval函数
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value //如果遇到了Return类型，则提早返回这个值
		case *object.Error:
			return result //异常处理
		}
	}

	return result
}

func evalBlockStatement(
	block *ast.BlockStatement,
	env *object.Environment,
) object.Object { //处理块
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env) //对每个语句求值

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result //返回
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(operator string, right object.Object) object.Object { //处理前缀表达式
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "&&":
		return nativeBoolToBooleanObject(left.ToBoolean() && right.ToBoolean())
	case operator == "||":
		return nativeBoolToBooleanObject(left.ToBoolean() || right.ToBoolean())
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}
func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalStringInfixExpression( //字符串比较
	operator string,
	left, right object.Object,
) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	switch operator {
	case "&&":
		return nativeBoolToBooleanObject(left.ToBoolean() && right.ToBoolean())
	case "||":
		return nativeBoolToBooleanObject(left.ToBoolean() || right.ToBoolean())
	case "+":
		return &object.String{Value: leftVal + rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "&&":
		return nativeBoolToBooleanObject(left.ToBoolean() && right.ToBoolean())
	case "||":
		return nativeBoolToBooleanObject(left.ToBoolean() || right.ToBoolean())
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIfExpression(
	ie *ast.IfExpression,
	env *object.Environment,
) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: " + node.Value)
}

func isTruthy(obj object.Object) bool { //参数是Boolean，返回布尔值
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalExpressions( //计算表达式，表达式可能有多个
	exps []ast.Expression,
	env *object.Environment,
) []object.Object { //可能返回多个值，将这些值存储在Object数组中返回
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object { //如果遇到函数调用，则直接执行该函数，如果有返回值，则返回它
	switch fn := fn.(type) {

	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		if result := fn.Fn(args...); result != nil {
			return result
		}
		return NULL

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func evalForExpression(fs *ast.ForExpression, env *object.Environment) object.Object {
	if !(fs.Initialize == nil) { //初始化
		Eval(fs.Initialize, env)
	}
	for isTruthy(Eval(fs.Condition, env)) {
		evaluated := Eval(fs.Body, env)

		// 检查循环体内语句执行后的返回值类型
		switch evaluated := evaluated.(type) {
		case *object.ReturnValue:
			return evaluated.Value
		case *object.BreakValue:
			// 当遇到 break 语句时，直接返回 NULL，即退出循环
			return NULL
		case *object.ContinueValue:
			// 当遇到 continue 语句时，重新评估条件，继续下一次循环
			continue
		}

		// 如果在循环体内遇到了 if 语句块，则需要对其进行评估
		if ifExpr, ok := fs.Body.Statements[len(fs.Body.Statements)-1].(*ast.IfExpression); ok {
			if isTruthy(Eval(ifExpr.Condition, env)) {
				// 如果条件满足，则执行 if 语句块的 Consequence
				evaluated := Eval(ifExpr.Consequence, env)
				switch evaluated := evaluated.(type) {
				case *object.ReturnValue:
					return evaluated.Value
				case *object.BreakValue:
					return NULL
				case *object.ContinueValue:
					continue
				}
			} else if ifExpr.Alternative != nil {
				// 如果条件不满足且存在 Alternative，则执行 Alternative
				evaluated := Eval(ifExpr.Alternative, env)
				switch evaluated := evaluated.(type) {
				case *object.ReturnValue:
					return evaluated.Value
				case *object.BreakValue:
					return nil
				case *object.ContinueValue:
					continue
				}
			}
		}
		//执行循环操作
		evaluated = Eval(fs.Cycleop, env)
		// 重新评估条件
		// 如果条件不再满足，则退出循环
		if !isTruthy(Eval(fs.Condition, env)) {
			break
		}
	}
	return NULL
}

func evalWhileExpression(fs *ast.WhileExpression, env *object.Environment) object.Object {
	for isTruthy(Eval(fs.Condition, env)) {
		evaluated := Eval(fs.Body, env)

		// 检查循环体内语句执行后的返回值类型
		switch evaluated := evaluated.(type) {
		case *object.ReturnValue:
			return evaluated.Value
		case *object.BreakValue:
			// 当遇到 break 语句时，直接返回 NULL，即退出循环
			return NULL
		case *object.ContinueValue:
			// 当遇到 continue 语句时，重新评估条件，继续下一次循环
			continue
		}

		// 如果在循环体内遇到了 if 语句块，则需要对其进行评估
		if ifExpr, ok := fs.Body.Statements[len(fs.Body.Statements)-1].(*ast.IfExpression); ok {
			if isTruthy(Eval(ifExpr.Condition, env)) {
				// 如果条件满足，则执行 if 语句块的 Consequence
				evaluated := Eval(ifExpr.Consequence, env)
				switch evaluated := evaluated.(type) {
				case *object.ReturnValue:
					return evaluated.Value
				case *object.BreakValue:
					return NULL
				case *object.ContinueValue:
					continue
				}
			} else if ifExpr.Alternative != nil {
				// 如果条件不满足且存在 Alternative，则执行 Alternative
				evaluated := Eval(ifExpr.Alternative, env)
				switch evaluated := evaluated.(type) {
				case *object.ReturnValue:
					return evaluated.Value
				case *object.BreakValue:
					return nil
				case *object.ContinueValue:
					continue
				}
			}
		}

		// 重新评估条件
		// 如果条件不再满足，则退出循环
		if !isTruthy(Eval(fs.Condition, env)) {
			break
		}
	}

	return NULL
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
		// hash表
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)
	if idx < 0 || idx > max {
		return NULL
	}
	return arrayObject.Elements[idx]
}

func evalHashLiteral(
	node *ast.HashLiteral,
	env *object.Environment,
) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

package evaluator

import (
	"my.com/myfile/object"
)

var builtins = map[string]*object.Builtin{
	"puts": object.GetBuiltinByName("puts"),
	"push": object.GetBuiltinByName("push"),
}

//var builtins = map[string]*object.Builtin{
//	"len": &object.Builtin{
//		Fn: func(args ...object.Object) object.Object {
//			if len(args) != 1 {
//				return newError("wrong number of arguments. got=%d,want=1")
//			}
//
//			switch arg := args[0].(type) {
//			case *object.String:
//				return &object.Integer{Value: int64(len(arg.Value))}
//			default:
//				return newError("arguments to `len` not supported, got  %s", args[0].Type())
//			}
//		},
//	},
//	"puts": &object.Builtin{
//		Fn: func(args ...object.Object) object.Object {
//			for _, arg := range args {
//				fmt.Print(arg.Inspect())
//			}
//
//			return nil
//		},
//	},
//	"first": &object.Builtin{
//		Fn: func(args ...object.Object) object.Object {
//			if len(args) != 1 {
//				return newError("wrong number of arguments. got=%d, want=1",
//					len(args))
//			}
//			if args[0].Type() != object.ARRAY_OBJ {
//				return newError("argument to `first` must be ARRAY, got %s",
//					args[0].Type())
//			}
//			arr := args[0].(*object.Array)
//			if len(arr.Elements) > 0 {
//				return arr.Elements[0]
//			}
//			return NULL
//		},
//	},
//	"last": &object.Builtin{
//		Fn: func(args ...object.Object) object.Object {
//			if len(args) != 1 {
//				return newError("wrong number of arguments. got=%d, want=1",
//					len(args))
//			}
//			if args[0].Type() != object.ARRAY_OBJ {
//				return newError("argument to `last` must be ARRAY, got %s",
//					args[0].Type())
//			}
//			arr := args[0].(*object.Array)
//			length := len(arr.Elements)
//			if length > 0 {
//				return arr.Elements[length-1]
//			}
//			return NULL
//		},
//	},
//	"rest": &object.Builtin{
//		Fn: func(args ...object.Object) object.Object {
//			if len(args) != 1 {
//				return newError("wrong number of arguments. got=%d, want=1",
//					len(args))
//			}
//			if args[0].Type() != object.ARRAY_OBJ {
//				return newError("argument to `rest` must be ARRAY, got %s",
//					args[0].Type())
//			}
//			arr := args[0].(*object.Array)
//			length := len(arr.Elements)
//			if length > 0 {
//				newElements := make([]object.Object, length-1, length-1)
//				copy(newElements, arr.Elements[1:length])
//				return &object.Array{Elements: newElements}
//			}
//			return NULL
//		},
//	},
//	"push": &object.Builtin{
//		Fn: func(args ...object.Object) object.Object {
//			// 对数组实现
//			if args[0].Type() == object.ARRAY_OBJ {
//				if len(args) != 2 {
//					return newError("wrong number of arguments. got=%d, want=2",
//						len(args))
//				}
//
//				arr := args[0].(*object.Array)
//				length := len(arr.Elements)
//				newElements := make([]object.Object, length+1, length+1)
//				copy(newElements, arr.Elements)
//				newElements[length] = args[1]
//				return &object.Array{Elements: newElements}
//
//			} else if args[0].Type() == object.HASH_OBJ {
//				if len(args) != 3 {
//					return newError("wrong number of arguments. got=%d, want=3",
//						len(args))
//				}
//
//				// 添加键值对
//				hashmap := args[0].(*object.Hash)
//				if hashmap.Pairs == nil {
//					return newError("hashmap not exist")
//				}
//				uintValue, _ := strconv.ParseUint(args[1].Inspect(), 10, 64)
//				key := object.HashKey{
//					Type:  args[1].Type(),
//					Value: uintValue,
//				}
//				value := args[2]
//				// TODO: 赋值
//				hashmap.Pairs[key] = object.HashPair{Key: args[1], Value: value}
//
//				return hashmap
//			} else {
//				return newError("argument to `push` must be ARRAY or HASHMAP, got %s",
//					args[0].Type())
//			}
//
//		},
//	},
//	"length": &object.Builtin{
//		Fn: func(args ...object.Object) object.Object {
//			// 只接受一个参赛，即要统计的数组
//			if len(args) != 1 {
//				return newError("wrong number of arguments. got=%d, want=2",
//					len(args))
//			}
//			// 只能对数组使用
//			if args[0].Type() != object.ARRAY_OBJ {
//				return newError("argument to `push` must be ARRAY, got %s",
//					args[0].Type())
//			}
//			arr := args[0].(*object.Array)
//			return &object.Integer{Value: int64(len(arr.Elements))}
//		},
//	},
//}

package object

import "fmt"

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"puts",
		&Builtin{Fn: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return nil
		},
		},
	},
	{
		Name: "push",
		Builtin: &Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 && len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=2 or 3", len(args))
			}

			if args[0].Type() == ARRAY_OBJ {
				if len(args) != 2 {
					return newError("wrong number of arguments for array push. got=%d, want=2", len(args))
				}
				arr := args[0].(*Array)
				length := len(arr.Elements)

				newElements := make([]Object, length+1, length+1)
				copy(newElements, arr.Elements)
				newElements[length] = args[1]

				return &Array{Elements: newElements}
			} else if args[0].Type() == HASH_OBJ {
				if len(args) != 3 {
					return newError("wrong number of arguments for hash push. got=%d, want=3", len(args))
				}
				hash := args[0].(*Hash)
				key := args[1]
				value := args[2]

				hashedKey := HashKey{
					Type: key.Type(),
				}

				newPairs := make(map[HashKey]HashPair)
				for k, v := range hash.Pairs {
					newPairs[k] = v
				}
				newPairs[hashedKey] = HashPair{Key: key, Value: value}

				return &Hash{Pairs: newPairs}
			} else {
				return newError("argument to `push` must be ARRAY or HASH, got %s", args[0].Type())
			}
		}},
	},
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func GetBuiltinByName(name string) *Builtin { // 通过名字获取内置函数
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}

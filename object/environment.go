package object //可以使用object.go中定义的Object接口

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

type Environment struct { //使用链表的结构来存储变量，使用Object接口来表示变量
	store map[string]Object
	outer *Environment
}

func (e *Environment) Get(name string) (Object, bool) { //Get方法能够
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

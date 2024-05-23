# hello Wizard
该编译器的结构如下：

* main:      程序开始，调用repl的Start函数开始编译
* repl：     允许用户输入代码并且调用lexer得到一个词法分析器,调用parser得到抽象语法树，调用evaluator求值
* lexer:     New生成词法分析器;提供了生成Token的方法
* token:     定义Token结构体，Token类型，关键字;提供了匹配关键字的函数，
* parser:    nextToken调用lexer生成Token;使用pratte算法定义优先级;提供了生成抽象语法树的函数
* ast:       定义了抽象语法树的结构体，接口和方法
* evaluator: Eval()求值，定义了不同语法树的求值方法
* object:    定义了返回值的类型和方法

module repl

go 1.22.1

//require "my.com/myfile/token" v0.0.0
//replace my.com/myfile/token => ../token
//
//require "my.com/myfile/lexer" v0.0.0
//replace my.com/myfile/lexer => ../lexer
//
//require "my.com/myfile/parser" v0.0.0
//replace my.com/myfile/parser => ../parser

require (
	my.com/myfile/token v0.0.0
    my.com/myfile/lexer v0.0.0
    my.com/myfile/parser v0.0.0
    my.com/myfile/ast v0.0.0
    my.com/myfile/object v0.0.0
    my.com/myfile/compiler v0.0.0
    my.com/myfile/code v0.0.0
    my.com/myfile/vm v0.0.0
    my.com/myfile/evaluator v0.0.0
)

replace (
	my.com/myfile/token => ../token
    my.com/myfile/lexer => ../lexer
    my.com/myfile/parser => ../parser
    my.com/myfile/ast => ../ast
    my.com/myfile/object => ../object
    my.com/myfile/vm => ../vm
    my.com/myfile/code => ../code
    my.com/myfile/compiler => ../compiler
    my.com/myfile/evaluator => ../evaluator
)


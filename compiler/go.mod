module compiler

go 1.22.1

require (
	my.com/myfile/token v0.0.0
	my.com/myfile/code v0.0.0
	my.com/myfile/lexer v0.0.0
    my.com/myfile/ast v0.0.0
    my.com/myfile/object v0.0.0
)

replace (
	my.com/myfile/token => ../token
	my.com/myfile/code  => ../code
	my.com/myfile/lexer => ../lexer
    my.com/myfile/ast => ../ast
    my.com/myfile/object => ../object
)
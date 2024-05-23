package parser

import (
	"fmt"
	"strconv"

	"my.com/myfile/ast"
	"my.com/myfile/lexer"
	"my.com/myfile/token"
)

// 定义优先级
const (
	_           int = iota //_ int = iota 表示从0开始自增
	LOWEST                 //
	EQUALS                 // ==
	LESSGREATER            // > or < or <= or >=
	LOGIGACLOR             //||
	LOGIGALAND             //&&
	SUM                    // +
	PRODUCT                // *
	PREFIX                 // -X or !X
	CALL                   // myFunction(X)
	INDEX
)

// 让优先级与token类型相匹配
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LE:       LESSGREATER,
	token.GE:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
	token.OR:       LOGIGACLOR,
	token.AND:      LOGIGALAND,
}

type (
	prefixParseFn func() ast.Expression               //prefixParseFn没有参数并返回ast.Expression类型的值
	infixParseFn  func(ast.Expression) ast.Expression //infixParseFn有参数并且返回ast.Expression类型的值
) //其中ast.Expression是一个接口

type Parser struct { //parser结构体
	l      *lexer.Lexer //指向Lexer结构体的指针
	errors []string     //储存解析过程中的错误信息

	curToken  token.Token //当前的token
	peekToken token.Token //预览token来进一步判断

	prefixParseFns map[token.TokenType]prefixParseFn //储存前缀表达式相关的解析函数
	infixParseFns  map[token.TokenType]infixParseFn  //...后缀...
}

func New(l *lexer.Lexer) *Parser { //返回一个parser结构体
	p := &Parser{ //定义p为一个Parser结构体
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn) //make()函数被用来创建一个空的映射，其中键的类型是token.TokenType，值的类型是prefixParseFn
	p.registerPrefix(token.ID, p.parseIdentifier)              //定义了对于不同类型的Token需要使用怎样的解析函数，ID的解析函数
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.WHILE, p.parseWhileExpression)
	p.registerPrefix(token.FOR, p.parserForExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LE, p.parseInfixExpression)
	p.registerInfix(token.GE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	//p.registerInfix(token.ASSIGN, p.parseAssignExpression)

	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken      //让curToken往前移
	p.peekToken = p.l.NextToken() //调用了lexer的NextToken方法，不准确地说，就是通过p访问l，再通过l的NextToken方法创建token，使用这种方法tokens不会保留
}

func (p *Parser) curTokenIs(t token.TokenType) bool { //匹配token类型
	return p.curToken.Type == t //如果curToken的类型是t类型，那么返回true，不是则返回false
}

func (p *Parser) peekTokenIs(t token.TokenType) bool { //同上，只不过是匹配之后一个Token
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool { //匹配token类型
	if p.peekTokenIs(t) { //与peekTokenIs不同，这个函数在匹配成功的时候不仅将返回布尔值，而且会next到下一个token
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// 下面定义了几种类型的错误
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}              //创建一个Program结构体
	program.Statements = []ast.Statement{} //将数组Statement作为ast.Statement数组

	for !p.curTokenIs(token.EOF) { //直到p的curtoken是EOF类型，执行for循环，这个循环的目的是生成所有的token，同时生成语法树
		stmt := p.parseStatement()
		if stmt != nil { //在循环的1，2次，p.curtoken为空，所以stmt没有被分配
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement { //判断应该返回什么类型的ast结构体
	switch p.curToken.Type { //提供了所有能够生成的根抽象语法树
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.CONTINUE:
		return p.parseContinueStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken, Index: ""} //p.curToken应该是Let

	if !p.expectPeek(token.ID) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal} //ID名称
	if p.peekTokenIs(token.LBRACKET) {
		p.nextToken()
		if p.expectPeek(token.INT) {
			if p.peekTokenIs(token.RBRACKET) {
				stmt.Index = p.curToken.Literal
				p.nextToken()
			}
		}
	}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST) //ID的值

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement { //返回语句的解析函数
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
func (p *Parser) parseBreakStatement() *ast.BreakStatement { //跳出语句
	stmt := &ast.BreakStatement{Token: p.curToken}

	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}

	return stmt
}

func (p *Parser) parseContinueStatement() *ast.ContinueStatement { //继续语句
	stmt := &ast.ContinueStatement{Token: p.curToken}

	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement { //创建表达式结构体
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) { //匹配';'，而在parseExpression中，虽然也能够匹配';'但是不会跳过，而且由于会比较优先级，也不是必须匹配
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression { //非常重要的函数，创建表达式的抽象语法树，如果是表达式的开始，一般使用LOWEST优先级
	prefix := p.prefixParseFns[p.curToken.Type] //寻找相应类型的前缀解析函数
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type) //如果没有找到，则报告一个错误
		return nil
	}
	leftExp := prefix() //如果找到解析函数，则使用该解析函数

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() { //如果下一个Token不是';'，而且优先级大于参数
		infix := p.infixParseFns[p.peekToken.Type] //寻找中缀解析函数
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp) //如果找到解析函数，则将leftxp扩展
	}

	return leftExp
}

func (p *Parser) peekPrecedence() int { //返回下一个Token的优先级
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int { //返回当前Token的优先级
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) parseIdentifier() ast.Expression { //ID的解析函数
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression { //处理负数和逻辑非
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression { //扩展前缀表达式为中缀表达式
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence() //返回当前Token的优先级给parseExpression判断
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression { //处理布尔值
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression { //处理表达式有括号的情况
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression { //处理if语句
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement { //块的解析函数
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression { //处理函数的定义
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier { //处理函数的参数
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression { //处理参数
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) { //储存前缀的解析函数
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) { //储存中缀的解析函数
	p.infixParseFns[tokenType] = fn
}
func (p *Parser) parserForExpression() ast.Expression { //处理for循环
	exp := &ast.ForExpression{Token: p.curToken} //for

	if !p.peekTokenIs(token.COLON) { //匹配冒号
		//如果匹配失败，则初始化
		p.nextToken() //跳过for
		exp.Initialize = p.parseLetStatement()
		if !p.expectPeek(token.COLON) { //继续匹配冒号
			return nil
		}
	} else {
		exp.Initialize = nil
		p.nextToken() //跳过for
	}
	p.nextToken() //跳过':'

	exp.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}
	p.nextToken()
	exp.Cycleop = p.parseLetStatement()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	exp.Body = p.parseBlockStatement()

	return exp
}
func (p *Parser) parseWhileExpression() ast.Expression { //处理While循环
	exp := &ast.WhileExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) { //匹配括号
		return nil
	}

	p.nextToken()
	exp.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	exp.Body = p.parseBlockStatement()

	return exp
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(end) {
		return nil
	}
	return list
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}

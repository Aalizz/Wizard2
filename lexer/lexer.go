// lexer.go
package lexer

import (
	"bytes"

	"my.com/myfile/token"
)

type Lexer struct { //Lexer的主体
	input        string //所有int类型的成员都被自动初始化为0
	position     int    //正在获取的字符的位置
	readPosition int    //需要读取的字符的位置
	ch           byte   //正在处理的字符
}

func New(input string) *Lexer {
	l := &Lexer{input: input} //将input作为Lexer结构体的input初始化l
	l.readChar()              //next操作，使得position=0,readposition=1
	return l                  //返回一个Lexer结构体的指针
}

func (l *Lexer) NextToken() token.Token { //受parser.nextToken调用
	var tok token.Token

	l.skipWhitespace() //跳过空格，制表符，换行符

	switch l.ch { //运算符判断
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.LE, Literal: literal} //LE就是<=
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.GE, Literal: literal} //GE就是>=
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.AND, Literal: literal}
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.AND, Literal: literal}
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()       //
			tok.Type = token.LookupId(tok.Literal) //判断是否为关键字，若不是返回ID作为类型，若是返回关键字map中对应的类型
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			if l.ch == '.' {
				tok.Type = token.FLOAT
				tok.Literal += string(l.ch)
				l.readChar()
				tok.Literal += l.readNumber()
			}
			if l.ch == 'E' || l.ch == 'e' {
				tok.Literal += string(l.ch)
				l.readChar()
				if l.ch == '+' || l.ch == '-' {
					tok.Literal += string(l.ch)
					l.readChar()
				}
				tok.Literal += l.readNumber()
			}
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok //返回一个token
}

func (l *Lexer) skipWhitespace() { //跳过
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readChar() { //next操作
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) peekChar() byte { //peekChar读取当前字符的后一个字符，如果有则返回进一步判断
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readIdentifier() string { //修改，应该是接受字符数字和下划线
	position := l.position //保存初始位置
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool { //isLetter里面也接受下划线
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)} //用来处理token的值是一个字符串的情况
}

func (l *Lexer) readString() string {
	var escaped bool     //标记当前解析的字符是否是转义字符
	var out bytes.Buffer //使用bytes.Buffer类型的out变量来累积解析出来的字符，这样能高效地构建字符串

	for {
		l.readChar() //首先消耗一个字符
		if l.ch == '"' || l.ch == 0 {
			break
		}

		if escaped {
			switch l.ch {
			case 'n':
				out.WriteRune('\n')
			case 't':
				out.WriteRune('\t')
			default:
				out.WriteByte(l.ch)
			}
			escaped = false
		} else {
			if l.ch == '\\' {
				escaped = true
			} else {
				out.WriteByte(l.ch)
			}
		}
	}
	return out.String()
}

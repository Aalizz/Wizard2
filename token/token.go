package token

type TokenType string

type Token struct {
	Type    TokenType //token的类型
	Literal string    //token的值
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// 标识符+字面量
	ID    = "ID"
	INT   = "INT"
	FLOAT = "FLOAT"

	// 运算符
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	EQ       = "=="
	NOT_EQ   = "!="
	GE       = ">="
	LE       = "<="

	LT = "<"
	GT = ">"

	//逻辑运算符
	OR  = "||"
	AND = "&&"

	// 分隔符
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	LBRACKET = "["
	RBRACKET = "]"

	// 关键字
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
	PRINT    = "PRINT"
	WHILE    = "WHILE"
	STRING   = "STRING"
	FOR      = "for"
)

// 判断是否是关键字
var keywords = map[string]TokenType{
	"fn":       FUNCTION,
	"let":      LET,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"return":   RETURN,
	"print":    PRINT,
	"while":    WHILE,
	"continue": CONTINUE,
	"break":    BREAK,
	"for":      FOR,
}

// LookupId 查找关键字，如果不是关键字则返回ID
func LookupId(id string) TokenType {
	keywords := keywords
	if tok, ok := keywords[id]; ok {
		return tok
	}
	return ID
}

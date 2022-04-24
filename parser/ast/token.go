package ast

type Token string

const (
	COMMENT   Token = "comment"
	SYNTAX    Token = "syntax"
	MODEL     Token = "model"
	REST      Token = "rest"
	GRPC      Token = "grpc"
	WS        Token = "ws"
	IMPORT    Token = "import"
	RAW       Token = "raw"
	SEPARATOR Token = ";"
	DECORATOR Token = "@"
	LPAREN    Token = "("
	LBRACE    Token = "{"
	COMMA     Token = ","
	PERIOD    Token = "."
	RPAREN    Token = ")"
	RBRACE    Token = "}"
)

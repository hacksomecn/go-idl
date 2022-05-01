/*
 * The MIT License (MIT)
 *
 * Copyright © 2022 Hao Luo <haozzzzzzzz@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

// Package ast imitate go/token/token.go
package ast

import "fmt"

// Token is the set of lexical tokens of the go-idl programming language.
type Token string

func (m Token) String() string {
	return string(m)
}

// The list of tokens.
const (
	// Special tokens

	ILLEGAL Token = "ILLEGAL"
	EOF           = "EOF"
	COMMENT       = "COMMENT"
	NEWLINE       = "NEWLINE"

	//literalBeg
	// Identifiers and basic type literals
	// (these tokens stand for classes of literals)

	IDENT  = "IDENT"  // main, func name, different declare content
	INT    = "INT"    // 12345
	FLOAT  = "FLOAT"  // 123.45
	IMAG   = "IMAG"   // 123.45i
	CHAR   = "CHAR"   // 'a'
	STRING = "STRING" // "abc"
	//literalEnd

	//operatorBeg
	// Operators and delimiters

	ASSIGN = "=" // =

	LPAREN = "(" // (
	LBRACK = "[" // [
	LBRACE = "{" // {
	COMMA  = "," // ,
	PERIOD = "." // .

	RPAREN = ")" // )
	RBRACK = "]" // ]
	RBRACE = "}" // }
	//SEMICOLON // ;
	//COLON // :
	//operatorEnd

	//keywordBeg
	//Keywords

	SYNTAX    = "SYNTAX"  // syntax
	SERVICE   = "SERVICE" // service
	MODEL     = "MODEL"   // model
	REST      = "REST"    // rest
	GRPC      = "GRPC"    // grpc
	WS        = "WS"      // ws
	IMPORT    = "IMPORT"  // import
	RAW       = "RAW"     // raw
	DECORATOR = "@"       // @
	//keywordEnd
)

var keywords = map[string]Token{
	"syntax":  SYNTAX,
	"service": SERVICE,
	"model":   MODEL,
	"rest":    REST,
	"grpc":    GRPC,
	"ws":      WS,
	"import":  IMPORT,
	"raw":     RAW,
	"@":       DECORATOR,
}

var operators = map[rune]Token{
	'=': ASSIGN,

	'(': LPAREN,
	'[': LBRACK,
	'{': LBRACE,
	',': COMMA,
	'.': PERIOD,

	')': RPAREN,
	']': RBRACK,
	'}': RBRACE,
	//';': SEMICOLON,
	//':': COLON,
}

// LookupKeywordIdent maps an identifier to its keyword token or IDENT (if not a keyword).
func LookupKeywordIdent(ident string) Token {
	if tok, isKeyword := keywords[ident]; isKeyword {
		return tok
	}
	return IDENT
}

func LookupOperatorToken(ch rune) (token Token, exists bool) {
	token, exists = operators[ch]
	return
}

// BlockTokenPair block pair token
var BlockTokenPair = map[Token]Token{ // start_token -> close_token
	LPAREN: RPAREN, // (...)
	LBRACE: RBRACE, // {...}
}

// BlockTokenReversePair reverse block pair token
var BlockTokenReversePair = map[Token]Token{
	RPAREN: LPAREN, // ) -> (
	RBRACE: LBRACE, // } -> {
}

func IsBlockStartToken(token Token) (yes bool, endToken Token) {
	endToken, yes = BlockTokenPair[token]
	return
}

func IsBlockEndToken(token Token) (yes bool, startToken Token) {
	startToken, yes = BlockTokenReversePair[token]
	return
}

type TokenPos struct {
	FilePos *FilePos `json:"file_pos"` // file pos

	LineNo     int // line no
	LineOffset int // line offset
	Offset     int // char offset
}

func NewTokenPos(filePos *FilePos, charOffset int) *TokenPos {
	return &TokenPos{
		FilePos: filePos,
		Offset:  charOffset,
	}
}

func (m *TokenPos) String() string {
	return fmt.Sprintf("%s:L%d", m.FilePos.FilePath, m.LineNo)
}

// IsValid reports whether the position is valid.
func (m *TokenPos) IsValid() bool {
	return m != nil && m.Offset > 0
}

type TokenFile struct {
	Pos         *FilePos `json:"pos"`
	LineOffsets []int    `json:"line_offsets"` // line offsets
}

func NewTokenFile(pos *FilePos) *TokenFile {
	return &TokenFile{
		Pos:         pos,
		LineOffsets: make([]int, 0),
	}
}

func (m *TokenFile) AddLineOffset(offset int) {
	m.LineOffsets = append(m.LineOffsets, offset)
}

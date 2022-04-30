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

package ast

type Token string

const (
	COMMENT   Token = "comment"
	SYNTAX    Token = "syntax"
	SERVICE   Token = "service"
	MODEL     Token = "model"
	REST      Token = "rest"
	GRPC      Token = "grpc"
	WS        Token = "ws"
	IMPORT    Token = "import"
	RAW       Token = "raw"
	DECORATOR Token = "@"
	LPAREN    Token = "("
	LBRACE    Token = "{"
	COMMA     Token = ","
	PERIOD    Token = "."
	RPAREN    Token = ")"
	RBRACE    Token = "}"
)

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

	Offset int // char offset
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

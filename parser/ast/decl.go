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

import (
	"fmt"
)

type FilePos struct {
	Package  string `json:"package"`
	FileName string `json:"file_name"`
	Name     string `json:"name"`
	FilePath string `json:"file_path"`
}

// Decl definition declare
type Decl struct {
	Expr []byte    // expr string
	Pos  *TokenPos // declare TypePos
	End  *TokenPos // decl end
}

type IDecl interface {
	Help() string // declare syntax help text
}

// An Ident node represents an identifier.
type Ident struct {
	Pos  *TokenPos
	Name string
}

type BasicLit struct {
	Pos   *TokenPos // literal position
	Kind  Token     // token.INT, token.FLOAT, token.IMAG, token.CHAR, or token.STRING
	Value string    // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
}

func (m *BasicLit) End() *TokenPos {
	return NewTokenPos(m.Pos.FilePos, m.Pos.Offset+len(m.Value))
}

// A Comment node represents a single //-style or /*-style comment.
//
// The Text field contains the comment text without carriage returns (\r) that
// may have been present in the source. Because a comment's end position is
// computed using len(Text), the position reported by End() does not match the
// true source end position for comments containing carriage returns.
type Comment struct {
	Pos  *TokenPos // position of "/" starting the comment
	Text string    // comment text (excluding '\n' for //-style comments)
}

func (c *Comment) Start() *TokenPos { return c.Pos }
func (c *Comment) End() *TokenPos {
	return NewTokenPos(c.Pos.FilePos, int(c.Pos.Offset+len(c.Text)))
}

func (c *Comment) String() string {
	return fmt.Sprintf("%s %s", c.Pos.String(), c.Text)
}

// A CommentGroup represents a sequence of comments
// with no other tokens and no empty lines between.
//
type CommentGroup struct {
	List []*Comment // len(List) > 0
}

func (g *CommentGroup) Pos() *TokenPos { return g.List[0].Start() }
func (g *CommentGroup) End() *TokenPos { return g.List[len(g.List)-1].End() }

//type CommentDecl struct {
//	Decl
//}
//
//func (m *CommentDecl) Help() string {
//	return `format: // or /**/
//usage: // for single line comment and /**/ for multiple line comment`
//}

type AssignmentDecl struct {
	Decl
	Doc     *CommentGroup
	Comment *CommentGroup
	Tok     Token
	Spec    *ValueSpec
}

func (m *AssignmentDecl) Help() string {
	return `format: <KEY>=<Value: golang type>`
}

type ValueSpec struct {
	Name  *Ident
	Value *BasicLit
}

type ImportDecl struct {
	Decl
	Doc       *CommentGroup
	Tok       Token
	LparenPos *TokenPos
	RparenPos *TokenPos
	Specs     []*ImportSpec
}

func (m *ImportDecl) Help() string {
	return "format: import \"<golang module name>\""
}

// An ImportSpec node represents a single package import.
type ImportSpec struct {
	Doc     *CommentGroup // associated documentation; or nil
	Comment *CommentGroup // local package name (including "."); or nil
	Name    *Ident        // import path
	Path    *BasicLit     // line comments; or nil
}

type ModelDecl struct {
	Decl
	Doc     *CommentGroup
	Comment *CommentGroup
	Tok     Token
	Name    *Ident
	Spec    *ModelType
}

func (m *ModelDecl) Help() string {
	return `TODO` // TODO
}

type ServiceDecl struct {
	Decl
	Doc     *CommentGroup
	Comment *CommentGroup
	Name    *Ident
	Tok     Token
}

func (m *ServiceDecl) Help() string {
	return `TODO` // TODO
}

type RestDecl struct {
	Decl
	Name       *Ident
	HttpMethod *Ident
	StrUri     *BasicLit
	Uris       []string
	Req        IType
	Resp       IType
}

func (m *RestDecl) Help() string {
	return `TODO` // TODO
}

type RestReq struct {
	*ModelType
	Merged bool // header、uri、query、body merged as one model
}

type GrpcDecl struct {
	Decl
	Name *Ident
	Req  IType
	Resp IType
}

func (m *GrpcDecl) Help() string {
	return `grpc [Handler]{
	req {}
	resp {}
}` // TODO
}

type WsDecl struct {
	Decl
	Name      *Ident    `json:"name"`
	Direction *Ident    `json:"direction"` // UP、DOWN
	MsgKind   *ValueLit `json:"msg_kind"`
	Msg       IType     `json:"msg"`
}

func (m *WsDecl) Help() string {
	return `TODO` // TODO
}

type RawDecl struct {
	Decl
}

func (m *RawDecl) Help() string {
	return `format: 
raw {
	.... 
}
usage: all text in block will be copy to generated .go file` // TODO
}

type DecoratorDecl struct {
	Decl
}

func (m *DecoratorDecl) Help() string {
	return `format: @<DECORATOR_NAME> decorator content text`
}

var HttpMethods = []string{
	"GET",
	"HEAD",
	"POST",
	"PUT",
	"PATCH",
	"DELETE",
	"CONNECT",
	"OPTIONS",
	"TRACE",
	"ANY",
}

var HttpMethodMap = map[string]bool{}

var WsDirections = []string{
	"UP",
	"Down",
}

var WsDirectionMap = map[string]bool{}

func init() {
	for _, method := range HttpMethods {
		HttpMethodMap[method] = true
	}

	for _, direction := range WsDirections {
		WsDirectionMap[direction] = true
	}
}

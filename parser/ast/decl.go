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

type FilePos struct {
	Package  string `json:"package"`
	FileName string `json:"file_name"`
	Name     string `json:"name"`
	FilePath string `json:"file_path"`
	Start    int    `json:"start"` // start pos in file
	End      int    `json:"end"`   // end pos in file
}

// Decl definition declare
type Decl struct {
	Expr string    // expr string
	Pos  *TokenPos // declare pos
}

type IDecl interface {
	Help() string // declare syntax help text
}

type AssignmentDecl struct {
	Decl
}

func (m *AssignmentDecl) Help() string {
	return `format: <KEY>=<VALUE: golang type>`
}

type CommentDecl struct {
	Decl
}

func (m *CommentDecl) Help() string {
	return `format: // or /**/
usage: // for single line comment and /**/ for multiple line comment`
}

type ImportDecl struct {
	Decl
}

func (m *ImportDecl) Help() string {
	return "format: import \"<golang module name>\""
}

type DecoratorDecl struct {
	Decl
}

func (m *DecoratorDecl) Help() string {
	return `format: @<DECORATOR_NAME> decorator content text`
}

type ServiceDecl struct {
	Decl
}

func (m *ServiceDecl) Help() string {
	return `TODO` // TODO
}

type ModelDecl struct {
	Decl
}

func (m *ModelDecl) Help() string {
	return `TODO` // TODO
}

type RestDecl struct {
	Decl
}

func (m *RestDecl) Help() string {
	return `TODO` // TODO
}

type GrpcDecl struct {
	Decl
}

func (m *GrpcDecl) Help() string {
	return `TODO` // TODO
}

type WsDecl struct {
	Decl
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

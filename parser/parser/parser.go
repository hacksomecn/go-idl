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

package parser

import (
	"fmt"
	"github.com/hacksomecn/go-idl/parser/ast"
	"github.com/hacksomecn/go-idl/parser/scanner"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strconv"
	"strings"
	"unicode"
)

type Parser struct {
	file    *ast.TokenFile
	idlFile *ast.IdlFile
	src     []byte

	errors  scanner.ErrorList
	scanner *scanner.Scanner

	// Comments
	comments        []*ast.CommentGroup
	lastLeadComment *ast.CommentGroup // last lead comment, not current
	lastLineComment *ast.CommentGroup // last line comment, not current

	// Next token, except comment
	pos *ast.TokenPos // token position
	tok ast.Token     // one token look-ahead
	lit string        // token literal

}

func NewParser(
	file *ast.TokenFile,
) (parser *Parser, err error) {
	parser = &Parser{
		file:    file,
		idlFile: ast.NewIdlFile(file),
	}
	err = parser.init()
	if err != nil {
		logrus.Errorf("init parser failed. error: %s", err)
		return
	}
	return
}

func (m *Parser) init() (err error) {
	src, err := ioutil.ReadFile(m.file.Pos.FilePath)
	if err != nil {
		logrus.Errorf("read file failed. error: %s", err)
		return
	}

	m.src = src
	m.scanner, err = scanner.NewScanner(m.file, src)
	if err != nil {
		logrus.Errorf("new scanner failed. error: %s", err)
		return
	}

	m.next()

	return
}

// Advance to the next non-comment token. In the process, collect
// any comment groups encountered, and remember the last lead and
// line comments.
//
// A lead comment is a comment group that starts and ends in a
// line without any other tokens and that is followed by a non-comment
// token on the line immediately after the comment group.
//
// A line comment is a comment group that follows a non-comment
// token on the same line, and that has no tokens after it on the line
// where it ends.
//
// Lead and line comments may be considered documentation that is
// stored in the AST.
//
func (m *Parser) next() {
	m.lastLeadComment = nil
	m.lastLineComment = nil
	prev := m.pos
	_ = prev
	m.next0()

	//read comment
	if m.tok == ast.COMMENT {
		var comment *ast.CommentGroup
		var endline int

		if m.pos.LineNo == prev.LineNo { // same line
			// The comment is on same line as the previous token; it
			// cannot be a lead comment but may be a line comment.
			comment, endline = m.consumeCommentGroup(0) // read one line
			if m.pos.LineNo != endline || m.tok == ast.EOF {
				// The next token is on a different line, thus
				// the last comment group is a line comment.
				m.lastLineComment = comment
			}
		}

		// consume successor comments, if any
		endline = -1
		for m.tok == ast.COMMENT { // iterate left comment, and get last one
			comment, endline = m.consumeCommentGroup(1)
		}

		// lead comment before current token 1 line
		if endline+1 == m.pos.LineNo {
			// The next token is following on the line immediately after the
			// comment group, thus the last comment group is a lead comment.
			m.lastLeadComment = comment
		}
	}
}

func (m *Parser) next0() {
	m.pos, m.tok, m.lit = m.scanner.Scan()
}

func (m *Parser) error(pos *ast.TokenPos, msg string) {
	m.errors.Add(pos, msg)
}

func (m *Parser) errorf(pos *ast.TokenPos, msg string, args ...any) {
	m.error(pos, fmt.Sprintf(msg, args...))
}

func (m *Parser) errorExpected(pos *ast.TokenPos, msg string) {
	msg = "expected " + msg
	m.error(pos, msg)
}

func (m *Parser) expect(tok ast.Token) *ast.TokenPos {
	pos := m.pos
	if m.tok != tok {
		m.errorExpected(pos, "'"+tok.String()+"'")
	}
	m.next() // make progress
	return pos
}

// expect2 is like expect, but it returns an invalid position
// if the expected token is not found.
func (m *Parser) expect2(tok ast.Token) (pos *ast.TokenPos) {
	if m.tok == tok {
		pos = m.pos
	} else {
		m.errorExpected(m.pos, "'"+tok.String()+"'")
	}
	m.next() // make progress
	return
}

// Consume a comment and return it and the line on which it ends.
func (m *Parser) consumeComment() (comment *ast.Comment, endline int) {
	// /*-style comments may end on a different line than where they start.
	// Scan the comment for '\n' chars and adjust endline accordingly.
	endline = m.pos.LineNo
	if m.lit[1] == '*' {
		// don't use range here - no need to decode Unicode code points
		for i := 0; i < len(m.lit); i++ {
			if m.lit[i] == '\n' {
				endline++
			}
		}
	}

	comment = &ast.Comment{Pos: m.pos, Text: m.lit}
	m.next0()
	return
}

// Consume a group of adjacent comments, add it to the parser's
// comments list, and return it together with the line at which
// the last comment in the group ends. A non-comment token or n
// empty lines terminate a comment group.
//
func (m *Parser) consumeCommentGroup(n int) (comments *ast.CommentGroup, endline int) {
	var list []*ast.Comment
	endline = m.pos.LineNo
	for m.tok == ast.COMMENT && m.pos.LineNo <= endline+n {
		var comment *ast.Comment
		comment, endline = m.consumeComment()
		list = append(list, comment)
	}

	// add comment group to the comments list
	comments = &ast.CommentGroup{List: list}
	m.comments = append(m.comments, comments)

	return
}

func (m *Parser) parseFile() (idlFile *ast.IdlFile) {
	var declHandlers = map[ast.Token]declHandlerFunc{
		ast.IMPORT:    m.parseImport,
		ast.SYNTAX:    m.parseAssignment,
		ast.SERVICE:   m.parseService,
		ast.MODEL:     m.parseModel,
		ast.REST:      m.parseRest,
		ast.GRPC:      m.parseGrpc,
		ast.WS:        m.parseWs,
		ast.RAW:       m.parseRaw,
		ast.DECORATOR: m.parseDecorator,
	}

	for m.tok != ast.EOF {
		//fmt.Println(m.pos, m.tok, m.lit)
		declHandler, ok := declHandlers[m.tok]
		if !ok {
			m.next()
			continue
		}

		declHandler()
	}
	return m.idlFile
}

type declHandlerFunc func() (decl ast.IDecl)

func (m *Parser) parseImport() (decl ast.IDecl) {
	doc := m.lastLeadComment
	pos := m.expect(ast.IMPORT)
	var lparent, rparen *ast.TokenPos
	specs := make([]*ast.ImportSpec, 0)
	exprStart := pos.Offset
	if doc != nil {
		exprStart = doc.Pos().Offset
	}

	var exprEnd int
	if m.tok == ast.LPAREN {
		lparent = m.pos
		//read import specs
		m.next()
		for i := 0; m.tok != ast.RPAREN && m.tok != ast.EOF; i++ {
			spec := m.parseImportSpec()
			specs = append(specs, spec)
		}
		rparen = m.expect(ast.RPAREN)
		exprEnd = rparen.Offset

	} else {
		// one spec
		spec := m.parseImportSpec()
		specs = append(specs, spec)
		// find line end
		if m.lastLineComment != nil {
			exprEnd = m.lastLineComment.End().Offset
		} else {
			exprEnd = spec.Path.Pos.Offset + len(spec.Path.Value)
		}
	}

	importDecl := &ast.ImportDecl{
		Decl: ast.Decl{
			Expr: m.src[exprStart : exprEnd+1],
			Pos:  pos,
		},
		Doc:       doc,
		Tok:       ast.IMPORT,
		LparenPos: lparent,
		RparenPos: rparen,
		Specs:     specs,
	}

	m.idlFile.AddImport(importDecl)

	return importDecl
}

func (m *Parser) parseImportSpec() (spec *ast.ImportSpec) {
	var ident *ast.Ident
	switch m.tok {
	case ast.PERIOD:
		ident = &ast.Ident{
			Pos:  m.pos,
			Name: ".",
		}
	case ast.IDENT:
		ident = m.parseIdent()
	}

	pos := m.pos
	var path string
	if m.tok == ast.STRING {
		path = m.lit
		if !isValidImport(path) {
			m.errorf(pos, "invalid import path: %s", path)
		}
		m.next()
	} else {
		m.expect(ast.STRING)
	}

	spec = &ast.ImportSpec{
		Doc:     m.lastLeadComment,
		Comment: m.lastLineComment,
		Name:    ident,
		Path: &ast.BasicLit{
			Pos:   pos,
			Kind:  ast.STRING,
			Value: path,
		},
	}
	return
}

func isValidImport(lit string) bool {
	const illegalChars = `!"#$%&'()*,:;<=>?[\]^{|}` + "`\uFFFD"
	s, _ := strconv.Unquote(lit) // go/scanner returns a legal string literal
	for _, r := range s {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) || strings.ContainsRune(illegalChars, r) {
			return false
		}
	}
	return s != ""
}

// ----------------------------------------------------------------------------
// Identifiers

func (m *Parser) parseIdent() *ast.Ident {
	pos := m.pos
	name := "_"
	if m.tok == ast.IDENT {
		name = m.lit
		m.next()
	} else {
		m.expect(ast.IDENT) // use expect() error handling
	}
	return &ast.Ident{Pos: pos, Name: name}
}

func (m *Parser) parseAssignment() (decl ast.IDecl) {
	// TODO temp
	m.next()
	return
}

func (m *Parser) parseService() (decl ast.IDecl) {
	// TODO temp
	m.next()
	return
}

func (m *Parser) parseModel() (decl ast.IDecl) {
	// TODO temp
	m.next()
	return
}

func (m *Parser) parseRest() (decl ast.IDecl) {
	// TODO temp
	m.next()
	return
}

func (m *Parser) parseGrpc() (decl ast.IDecl) {
	// TODO temp
	m.next()
	return
}

func (m *Parser) parseWs() (decl ast.IDecl) {
	// TODO temp
	m.next()
	return
}

func (m *Parser) parseRaw() (decl ast.IDecl) {
	// TODO temp
	m.next()
	return
}

func (m *Parser) parseDecorator() (decl ast.IDecl) {
	// TODO temp
	m.next()
	return
}

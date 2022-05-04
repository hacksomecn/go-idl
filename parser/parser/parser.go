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
	"unicode/utf8"
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

	// Error recovery
	// (used to limit the number of calls to parser.advance
	// w/o making scanning progress - avoids potential endless
	// loops across multiple parser functions during error recovery)
	syncPos ast.TokenPos // last synchronization position
	syncCnt int          // number of parser.advance calls without progress

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

	m.pos = &ast.TokenPos{
		FilePos: m.file.Pos,
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
		m.errorf(pos, "expected %q, but got %q", tok.String(), m.tok.String())
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

func (m *Parser) printToken0() {
	for m.tok != ast.EOF {
		// debug
		fmt.Println(m.pos, m.tok, m.lit)
		m.next0()
	}
}

func (m *Parser) parseFile() (idlFile *ast.IdlFile) {
	var declHandlers = map[ast.Token]declHandlerFunc{
		ast.IMPORT:    m.parseImport,
		ast.SYNTAX:    m.parseAssignment,
		ast.MODEL:     m.parseModel,
		ast.SERVICE:   m.parseService,
		ast.REST:      m.parseRest,
		ast.GRPC:      m.parseGrpc,
		ast.WS:        m.parseWs,
		ast.RAW:       m.parseRaw,
		ast.DECORATOR: m.parseDecorator,
	}
	_ = declHandlers

	for m.tok != ast.EOF {
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

	var exprEnd int // contain comment
	var declEnd int
	if m.tok == ast.LPAREN {
		lparent = m.pos
		//read import specs
		m.next()
		for i := 0; m.tok != ast.RPAREN && m.tok != ast.EOF; i++ {
			spec := m.parseImportSpec()
			specs = append(specs, spec)
		}
		rparen = m.expect(ast.RPAREN)
		m.expectSemi()
		exprEnd = rparen.Offset
		declEnd = rparen.Offset
	} else {
		// one spec
		spec := m.parseImportSpec()
		specs = append(specs, spec)
		// find line end
		if m.lastLineComment != nil {
			exprEnd = m.lastLineComment.End().Offset
		} else {
			exprEnd = spec.Path.Pos.Offset + len(spec.Path.Value)
			declEnd = spec.Path.Pos.Offset + len(spec.Path.Value)
		}
	}

	importDecl := &ast.ImportDecl{
		Decl: ast.Decl{
			Expr: m.src[exprStart : exprEnd+1],
			Pos:  pos,
			End:  ast.NewTokenPos(pos.FilePos, declEnd),
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

	m.expectSemi()

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
	pos := m.pos
	lit := m.lit
	tok := m.tok
	doc := m.lastLeadComment

	m.next()
	m.expect(ast.ASSIGN)

	var value *ast.BasicLit
	if IsBasicLitKind(m.tok) {
		value = &ast.BasicLit{
			Pos:   m.pos,
			Kind:  m.tok,
			Value: m.lit,
		}
	} else {
		value = &ast.BasicLit{}
		m.errorf(m.pos, "illegal basic type value. tok: %s", m.tok)
	}

	m.next()
	m.expectSemi() // call before accessing m.lastLineComment

	comment := m.lastLeadComment

	exprStart := pos.Offset
	if doc != nil {
		exprStart = doc.Pos().Offset
	}

	exprEnd := value.End().Offset
	if comment != nil {
		exprEnd = comment.End().Offset
	}

	assignDecl := &ast.AssignmentDecl{
		Decl: ast.Decl{
			Expr: m.src[exprStart : exprEnd+1],
			Pos:  pos,
			End:  ast.NewTokenPos(pos.FilePos, value.Pos.Offset+len(value.Value)),
		},
		Doc:     m.lastLeadComment,
		Comment: m.lastLineComment,
		Tok:     tok,
		Spec: &ast.ValueSpec{
			Name: &ast.Ident{
				Pos:  pos,
				Name: lit,
			},
			Value: value,
		},
	}
	m.idlFile.AddAssign(assignDecl)
	return assignDecl
}

func IsBasicLitKind(tok ast.Token) bool {
	switch tok {
	case ast.INT,
		ast.FLOAT,
		ast.IMAG,
		ast.CHAR,
		ast.STRING:
		return true
	}
	return false
}

func (m *Parser) parseModel() (decl ast.IDecl) {
	pos := m.pos
	//lit := m.lit
	tok := m.tok

	exprStart := pos.Offset
	doc := m.lastLeadComment
	comment := m.lastLineComment
	if doc != nil {
		exprStart = doc.Pos().Offset
	}

	m.next()
	// parse type
	spec := m.parseNamedModelSpec()
	m.expectSemi() // independent model finish with };

	exprEnd := spec.Closing.Offset
	comment = m.lastLineComment
	if comment != nil {
		exprEnd = comment.End().Offset
	}

	spec.Doc = doc
	modelDecl := &ast.ModelDecl{
		Decl: ast.Decl{
			Expr: m.src[exprStart : exprEnd+1],
			Pos:  pos,
			End:  spec.Closing,
		},
		Doc:     doc,
		Comment: comment,
		Tok:     tok,
		Name:    spec.Name,
		Spec:    spec,
	}
	m.idlFile.AddModel(modelDecl)
	return modelDecl
}

// required name
func (m *Parser) parseNamedModelSpec() (spec *ast.ModelType) {
	pos := m.pos
	name := m.parseIdent()
	spec = m.parseModelSpec()
	spec.Name = name
	spec.TypePos = pos
	return
}

// optional
func (m *Parser) parseModelSpecAnonymous() (spec *ast.ModelType) {
	spec = m.parseModelSpec()
	spec.Anonymous = true
	return
}

func (m *Parser) parseType() (iType ast.IType) {
	switch m.tok {
	case ast.IDENT:
		iType = m.parseTypeName(nil)
		return

	case ast.LBRACE: // { anonymous struct
		iType = m.parseModelSpecAnonymous()
		return

	case ast.LBRACK: // array
		iType = m.parseArrayType()
		return

	case ast.MAP: // map
		return m.parseMapType()

	case ast.INTERFACE: // interface
		return m.parseInterfaceType()

	case ast.Star:
		return m.parsePointerType()

	default:
		iType = ast.UnknownType
	}

	return
}

func (m *Parser) parseTypeName(
	ident *ast.Ident,
) (
	iType ast.IType,
) {
	var pos *ast.TokenPos
	if ident == nil {
		pos = m.pos
		ident = m.parseIdent()
	} else {
		pos = ident.Pos
	}

	if m.tok == ast.PERIOD {
		// ident is a package name
		m.next()
		sel := m.parseIdent()
		fullName := fmt.Sprintf("%s.%s", ident.Name, sel.Name)
		typeRef := &ast.TypeRef{
			Type: ast.Type{
				Name: &ast.Ident{
					Pos:  pos,
					Name: fullName,
				},
				TypePos: pos,
				TypeEnd: ast.NewTokenPos(pos.FilePos, pos.Offset+len(fullName)),
			},
			Package: ident,
			RefType: nil,
		}
		iType = typeRef
		return
	}

	typeRef := &ast.TypeRef{
		Type: ast.Type{
			Name:    ident,
			TypePos: pos,
			TypeEnd: ast.NewTokenPos(pos.FilePos, pos.Offset+len(ident.Name)),
		},
		Package: nil,
		RefType: nil,
	}
	iType = typeRef
	return
}

func (m *Parser) parseModelSpec() (spec *ast.ModelType) {
	doc := m.lastLeadComment
	lbrace := m.expect(ast.LBRACE)
	comment := m.lastLineComment
	fields := make([]*ast.ModelField, 0)

	// TypeNameIdent
	// *TypeNameIdent
	// Field TypeNameIdent
	// Field *TypeNameIdent
	for m.tok == ast.IDENT || m.tok == ast.Star {
		fields = append(fields, m.parseField())
	}

	rbrace := m.expect(ast.RBRACE)

	spec = &ast.ModelType{
		Type: ast.Type{
			Name: &ast.Ident{
				Pos:  lbrace,
				Name: "model",
			},
			TypePos: lbrace,
			TypeEnd: rbrace,
		},
		Doc:     doc,
		Comment: comment,
		Opening: lbrace,
		Fields:  fields,
		Closing: rbrace,
	}

	return
}

func (m *Parser) parseField() (field *ast.ModelField) {
	doc := m.lastLeadComment
	pos := m.pos

	var typ ast.IType
	var name *ast.Ident
	embedded := false
	if m.tok == ast.IDENT { // field name or embedded type
		name = m.parseIdent()
		if m.tok == ast.PERIOD || // package module
			m.tok == ast.STRING || // field tag
			m.tok == ast.SEMICOLON || // end declare
			m.tok == ast.RBRACE { // end type
			typ = m.parseTypeName(name)
			embedded = true
		} else {
			typ = m.parseType()
		}

	} else if m.tok == ast.Star {
		typ = m.parsePointerType()
		name = typ.TypeNameIdent()
		embedded = true
	} else {
		m.errorf(pos, "unsupported tok for field. %s", m.tok)
	}

	var tag *ast.FieldTag
	if m.tok == ast.STRING {
		tag = &ast.FieldTag{
			BasicLit: &ast.BasicLit{
				Pos:   m.pos,
				Kind:  m.tok,
				Value: m.lit,
			},
		}
		m.next()
	}

	m.expectSemi() // end expression, call before accessing p.linecomment

	field = &ast.ModelField{
		Pos:      pos,
		Doc:      doc,
		Comment:  m.lastLineComment,
		Name:     name,
		Type:     typ,
		Tag:      tag,
		Exported: IsExported(name.Name),
		Embedded: embedded,
	}
	return
}

func (m *Parser) expectSemi() {
	// semicolon is optional before a closing ')' or '}'
	if m.tok != ast.RPAREN && m.tok != ast.RBRACE {
		switch m.tok {
		case ast.COMMA:
			// permit a ',' instead of a ';' but complain
			m.errorExpected(m.pos, "';'")
			fallthrough
		case ast.SEMICOLON:
			m.next()
		default:
			m.errorExpected(m.pos, "';'")
			//m.advance(stmtStart)
		}
	}
}

// IsExported reports whether name starts with an upper-case letter.
//
func IsExported(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(ch)
}

// []T
func (m *Parser) parseArrayType() (arrayType *ast.ArrayType) {
	pos := m.pos
	lbrack := m.expect(ast.LBRACK)
	// not support [len] or [...] yet
	m.expect(ast.RBRACK)
	elt := m.parseType()
	arrayType = &ast.ArrayType{
		Type: ast.Type{
			Name: &ast.Ident{
				Pos:  pos,
				Name: fmt.Sprintf("[]%s", elt.TypeNameIdent().Name),
			},
			TypePos: pos,
			TypeEnd: ast.NewTokenPos(pos.FilePos, lbrack.Offset+elt.End().Offset),
		},
		ElemType: elt,
	}
	return
}

func (m *Parser) parseMapType() (mapType *ast.MapType) {
	pos := m.expect(ast.MAP)
	m.expect(ast.LBRACK)
	key := m.parseType()
	m.expect(ast.RBRACK)
	elt := m.parseType()

	mapType = &ast.MapType{
		Type: ast.Type{
			Name: &ast.Ident{
				Pos:  pos,
				Name: fmt.Sprintf("map[%s]%s", key.TypeNameIdent().Name, elt.TypeNameIdent().Name),
			},
			TypePos: pos,
			TypeEnd: elt.End(),
		},
		KeyType:  key,
		ElemType: elt,
	}
	return
}

// interface or interface{}
func (m *Parser) parseInterfaceType() (interfaceType *ast.InterfaceType) {
	pos := m.expect(ast.INTERFACE)
	end := ast.NewTokenPos(pos.FilePos, pos.Offset+len(m.lit))
	if m.tok == ast.LBRACE { // interface{}
		m.next()
		end = m.expect(ast.RBRACE)
	}

	interfaceType = &ast.InterfaceType{
		Type: ast.Type{
			Name: &ast.Ident{
				Pos:  pos,
				Name: "interface",
			},
			TypePos: pos,
			TypeEnd: end,
		},
	}
	return
}

func (m *Parser) parsePointerType() (pointerType *ast.PointerType) {
	pos := m.expect(ast.Star)
	base := m.parseType()

	pointerType = &ast.PointerType{
		Type: ast.Type{
			Name: &ast.Ident{
				Pos:  pos,
				Name: fmt.Sprintf("*%s", base.TypeNameIdent().Name),
			},
			TypePos: pos,
			TypeEnd: base.End(),
		},
		BaseType: base,
	}
	return
}

func (m *Parser) parseService() (decl ast.IDecl) {
	doc := m.lastLeadComment
	pos := m.pos
	tok := m.tok
	m.next()
	name := m.parseIdent()

	m.expect(ast.LBRACE)
	for m.tok != ast.RBRACE {
		m.next() // remain
	}

	rbrace := m.expect(ast.RBRACE)
	m.expectSemi()
	comment := m.lastLineComment

	exprStart := pos.Offset
	exprEnd := rbrace.Offset
	if doc != nil {
		exprStart = doc.Pos().Offset
		exprEnd = comment.End().Offset
	}

	serviceDecl := &ast.ServiceDecl{
		Decl: ast.Decl{
			Expr: m.src[exprStart : exprEnd+1],
			Pos:  pos,
			End:  rbrace,
		},
		Doc:     doc,
		Comment: comment,
		Name:    name,
		Tok:     tok,
	}
	decl = serviceDecl
	m.idlFile.AddService(serviceDecl)
	return
}

func (m *Parser) parseRest() (decl ast.IDecl) {
	pos := m.pos
	doc := m.lastLeadComment
	m.next()
	name := m.parseIdent()
	httpMethod := m.parseIdent()

	if !ast.HttpMethodMap[httpMethod.Name] {
		m.errorf(httpMethod.Pos, "unknown http method %s. available: %s", httpMethod.Name, ast.HttpMethods)
	}

	if m.tok != ast.STRING {
		m.errorf(m.pos, "require uri string")
	}

	strUri := &ast.BasicLit{
		Pos:   m.pos,
		Kind:  ast.STRING,
		Value: m.lit,
	}
	strUris := strings.Split(m.lit, ",")
	for idx, strUri := range strUris {
		strUris[idx] = strings.TrimSpace(strUri)
	}

	m.next()
	m.expect(ast.LBRACE)

	var req ast.IType
	var resp ast.IType

	fieldExists := make(map[string]bool)
	for m.tok == ast.IDENT {
		restFieldName := m.lit
		if fieldExists[restFieldName] {
			m.errorf(m.pos, "duplicated rest field. %s", restFieldName)
			continue
		}

		fieldExists[restFieldName] = true

		m.next()
		switch restFieldName {
		case "req":
			req = m.parseType()

		case "resp":
			resp = m.parseType()

		default:
			m.errorf(m.pos, "unknown field for rest. name: %s", m.lit)

		}

		m.expectSemi()
	}

	rbrace := m.expect(ast.RBRACE)
	m.expectSemi()
	comment := m.lastLineComment

	exprStart := pos.Offset
	exprEnd := rbrace.Offset
	if doc != nil {
		exprStart = doc.Pos().Offset
	}

	if comment != nil {
		exprEnd = comment.End().Offset
	}

	restDecl := &ast.RestDecl{
		Decl: ast.Decl{
			Expr: m.src[exprStart : exprEnd+1],
			Pos:  pos,
			End:  rbrace,
		},
		Name:       name,
		HttpMethod: httpMethod,
		StrUri:     strUri,
		Uris:       strUris,
		Req:        req,
		Resp:       resp,
	}
	decl = restDecl
	m.idlFile.AddRest(restDecl)
	return
}

func (m *Parser) checkRestReq(reqType *ast.ModelType) (req *ast.RestReq) {
	req = &ast.RestReq{
		ModelType: reqType,
		Merged:    false,
	}

	hasMerged := false
	hasSeperated := false // seperated
	for _, field := range reqType.Fields {
		if !restReqFields[field.Name.Name] {
			hasMerged = true
		} else {
			hasSeperated = true
		}
	}

	if hasMerged && hasSeperated {
		m.errorf(reqType.Pos(), "rest req can not has merged fields while has [Header|Uri|Query|Body] fields")
		return
	}

	req.Merged = hasMerged

	return
}

var restReqFields = map[string]bool{
	"Header": true,
	"Uri":    true,
	"Query":  true,
	"Body":   true,
}

func (m *Parser) parseGrpc() (decl ast.IDecl) {
	pos := m.pos
	doc := m.lastLeadComment
	m.next()
	name := m.parseIdent()
	m.expect(ast.LBRACE)

	var req ast.IType
	var resp ast.IType

	fieldExists := make(map[string]bool)
	for m.tok == ast.IDENT {
		fieldName := m.lit
		if fieldExists[fieldName] {
			m.errorf(m.pos, "duplicated grpc field. %s", fieldName)
			continue
		}
		fieldExists[fieldName] = true

		m.next()
		switch fieldName {
		case "req":
			req = m.parseType()

		case "resp":
			resp = m.parseType()

		default:
			m.errorf(m.pos, "unknown field for grpc. name:%s", m.lit)
		}

		m.expectSemi()
	}

	rbrace := m.expect(ast.RBRACE)
	m.expectSemi()
	comment := m.lastLineComment

	exprStart := pos.Offset
	exprEnd := rbrace.Offset
	if doc != nil {
		exprStart = doc.Pos().Offset
	}

	if comment != nil {
		exprEnd = comment.End().Offset
	}

	grpcDecl := &ast.GrpcDecl{
		Decl: ast.Decl{
			Expr: m.src[exprStart : exprEnd+1],
			Pos:  pos,
			End:  rbrace,
		},
		Name: name,
		Req:  req,
		Resp: resp,
	}
	decl = grpcDecl
	m.idlFile.AddGrpc(grpcDecl)
	return
}

func (m *Parser) parseWs() (decl ast.IDecl) {
	pos := m.pos
	doc := m.lastLeadComment
	m.next()
	name := m.parseIdent()
	direction := m.parseIdent()

	if !ast.WsDirectionMap[direction.Name] {
		m.errorf(direction.Pos, "unknown ws message direction %s. available: %s", direction, ast.WsDirections)
	}

	msgKind := &ast.ValueLit{
		Pos:   m.pos,
		Tok:   m.tok,
		Value: m.lit,
	}
	if !IsBasicLitKind(m.tok) && m.tok != ast.IDENT {
		m.errorf(m.pos, "illegal ws msg code. tok: %s", m.tok)
	}

	m.next()

	msg := m.parseType()
	m.expectSemi()
	comment := m.lastLineComment

	exprStart := pos.Offset
	exprEnd := msg.End().Offset
	if doc != nil {
		exprStart = doc.Pos().Offset
	}

	if comment != nil {
		exprEnd = comment.End().Offset
	}

	wsDecl := &ast.WsDecl{
		Decl: ast.Decl{
			Expr: m.src[exprStart : exprEnd+1],
			Pos:  pos,
			End:  msg.End(),
		},
		Name:      name,
		Direction: direction,
		MsgKind:   msgKind,
		Msg:       msg,
	}
	decl = wsDecl
	m.idlFile.AddWs(wsDecl)

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

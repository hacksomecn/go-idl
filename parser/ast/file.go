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

type IdlFile struct {
	*TokenFile

	Assigns      []*AssignmentDecl `json:"assigns"` // idl property assignment
	Imports      []*ImportDecl     `json:"imports"`
	Models       []*ModelDecl      `json:"models"`
	Services     []*ServiceDecl    `json:"services"`
	Rests        []*RestDecl       `json:"rests"`
	Grpcs        []*GrpcDecl       `json:"grpcs"`
	Wss          []*WsDecl         `json:"wss"`
	Raws         []*RawDecl        `json:"raws"`
	CommentGroup CommentGroup      `json:"comment_group"`

	Decls []IDecl `json:"decls" ` // all decl in sequence
}

func NewIdlFile(tokenFile *TokenFile) (file *IdlFile) {
	return &IdlFile{
		TokenFile:    tokenFile,
		Assigns:      []*AssignmentDecl{},
		Imports:      []*ImportDecl{},
		Models:       []*ModelDecl{},
		Services:     []*ServiceDecl{},
		Rests:        []*RestDecl{},
		Grpcs:        []*GrpcDecl{},
		Wss:          []*WsDecl{},
		Raws:         []*RawDecl{},
		CommentGroup: CommentGroup{},
		Decls:        []IDecl{},
	}
}

func (m *IdlFile) AddAssign(assign *AssignmentDecl) {
	m.Assigns = append(m.Assigns, assign)
	m.Decls = append(m.Decls, assign)
}

func (m *IdlFile) AddImport(imp *ImportDecl) {
	m.Imports = append(m.Imports, imp)
	m.Decls = append(m.Decls, imp)
}

func (m *IdlFile) AddModel(model *ModelDecl) {
	m.Models = append(m.Models, model)
	m.Decls = append(m.Decls, model)
}

func (m *IdlFile) AddService(service *ServiceDecl) {
	m.Services = append(m.Services, service)
	m.Decls = append(m.Decls, service)
}

func (m *IdlFile) AddRest(rest *RestDecl) {
	m.Rests = append(m.Rests, rest)
	m.Decls = append(m.Decls, rest)
}

func (m *IdlFile) AddGrpc(grpc *GrpcDecl) {
	m.Grpcs = append(m.Grpcs, grpc)
	m.Decls = append(m.Decls, grpc)
}

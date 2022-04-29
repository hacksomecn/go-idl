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
	Pos *FilePos `json:"pos"` // file position

	Assigns map[*FilePos]*AssignmentDecl `json:"assigns"` // idl property assignment
	Import  map[*FilePos]*ImportDecl     `json:"imports"`
	Models  map[*FilePos]*ModelDecl      `json:"models"`
	Rests   map[*FilePos]*RestDecl       `json:"rests"`
	Grpcs   map[*FilePos]*GrpcDecl       `json:"grpcs"`
	Wss     map[*FilePos]*WsDecl         `json:"wss"`
	Raws    map[*FilePos]*RawDecl        `json:"raws"`

	Stmts []IDecl `json:"stmts"` // all decl in sequence
}

func NewIdlFile() (file *IdlFile) {
	return &IdlFile{
		Assigns: map[*FilePos]*AssignmentDecl{},
		Import:  map[*FilePos]*ImportDecl{},
		Models:  map[*FilePos]*ModelDecl{},
		Rests:   map[*FilePos]*RestDecl{},
		Grpcs:   map[*FilePos]*GrpcDecl{},
		Wss:     map[*FilePos]*WsDecl{},
		Raws:    map[*FilePos]*RawDecl{},
		Stmts:   []IDecl{},
	}
}

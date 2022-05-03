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

import "strings"

type IType interface {
	Pos() *TokenPos
	End() *TokenPos
}

type Type struct {
}

func (m *Type) Pos() *TokenPos {
	return &TokenPos{}
}

func (m *Type) End() *TokenPos {
	return &TokenPos{}
}

var UnknownType = &Type{}

type ModelType struct {
	TypePos *TokenPos
	Doc     *CommentGroup
	Comment *CommentGroup
	Name    *Ident

	Opening *TokenPos
	Fields  []*ModelField
	Closing *TokenPos
}

func (m *ModelType) Pos() *TokenPos {
	return m.TypePos
}

func (m *ModelType) End() *TokenPos {
	return m.Closing
}

type StructType = ModelType

type ModelField struct {
	Pos     *TokenPos
	Doc     *CommentGroup
	Comment *CommentGroup
	Name    *Ident
	Type    IType
	Tag     *FieldTag

	// Exported reports whether the object is exported (starts with a capital letter).
	// It doesn't take into account whether the object is in a local (function) scope
	// or not.
	Exported bool

	// Embedded reports whether the variable is an embedded field.
	Embedded bool
}

type FieldTag struct {
	*BasicLit
}

func splitTag(strTag string) (mParts map[string]string) {
	mParts = make(map[string]string, 0)
	tagValue := strings.Replace(strTag, "`", "", -1)
	strPairs := strings.Split(tagValue, " ")
	for _, pair := range strPairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		separateIndex := strings.Index(pair, ":")
		if separateIndex < 0 || separateIndex == len(pair)-1 {
			continue
		}

		key := pair[:separateIndex]
		value := pair[separateIndex+1:]

		mParts[key] = strings.Replace(value, "\"", "", -1)
	}
	return
}

// TypeRef Refer to type
type TypeRef struct {
	TypePos *TokenPos
	Name    *Ident
	Package *Ident
	Type    IType
}

func (m *TypeRef) Pos() *TokenPos {
	return m.TypePos
}

func (m *TypeRef) End() *TokenPos {
	return NewTokenPos(m.TypePos.FilePos, m.TypePos.Offset+len(m.Name.Name))
}

type ArrayType struct {
	TypePos  *TokenPos
	Name     *Ident
	ElemType IType
}

func (m *ArrayType) Pos() *TokenPos {
	return m.TypePos
}

func (m *ArrayType) End() *TokenPos {
	return m.ElemType.End()
}

type MapType struct {
	TypePos  *TokenPos
	Name     *Ident
	KeyType  IType
	ElemType IType
}

func (m *MapType) Pos() *TokenPos {
	return m.TypePos
}

func (m *MapType) End() *TokenPos {
	return m.ElemType.End()
}

type InterfaceType struct {
	TypePos *TokenPos
	Name    *Ident
}

func (m *InterfaceType) Pos() *TokenPos {
	return m.TypePos
}

func (m *InterfaceType) End() *TokenPos {
	return NewTokenPos(m.TypePos.FilePos, m.TypePos.Offset+len(m.Name.Name))
}

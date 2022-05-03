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

// Package scanner Go-idl scanner error imitates go/scanner/errors.go
package scanner

import (
	"fmt"
	"github.com/hacksomecn/go-idl/parser/ast"
)

type Error struct {
	Pos *ast.TokenPos
	Msg string
}

func (m Error) Error() (str string) {
	if m.Pos != nil {
		str = fmt.Sprintf("%s:LINE%d %s", m.Pos.FilePos.FileName, m.Pos.LineNo, m.Msg)
	} else {
		str = m.Msg
	}
	return
}

type ErrorList []*Error

func (m *ErrorList) Add(pos *ast.TokenPos, msg string) {
	*m = append(*m, &Error{Pos: pos, Msg: msg})
	return
}

func (m ErrorList) Error() (str string) {
	switch len(m) {
	case 0:
		return "no errors"
	case 1:
		return m[0].Error()
	}
	return fmt.Sprintf("%s and (%d) more errors", m[0], len(m)-1)
}

func (m ErrorList) Err() error {
	if len(m) == 0 {
		return nil
	}

	return m
}

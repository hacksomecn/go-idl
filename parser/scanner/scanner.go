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

// Package scanner implements a scanner for go-idl src text.
// It takes a []byte as src which can then be tokenized
// through repeated calls to the Scan method.
//
// Go-idl scanner imitated go/scanner.
package scanner

import (
	"bytes"
	"fmt"
	"github.com/hacksomecn/go-idl/gopkg"
	"github.com/hacksomecn/go-idl/parser/ast"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"unicode"
	"unicode/utf8"
)

// ScanFiles scan .gidl files in dir
func ScanFiles(
	path string, // file or dir path
	modulePackagePath string, // package path with module name, if empty will find package for path
) (
	files []*ast.TokenFile,
	fileMap map[string]*ast.TokenFile, // file abs path -> *ast.TokenFile
	err error,
) {
	// get package name
	fileNames, err := FindIdlFiles(path)
	if err != nil {
		logrus.Errorf("get idl files failed. error: %s", err)
		return
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		logrus.Errorf("get abs path failed. path: %s, error: %s", path, err)
		return
	}

	isAbsPathExists, isAbsPathADir := PathExists(absPath)
	if !isAbsPathExists {
		err = fmt.Errorf("file path not exists. %s", absPath)
		return
	}

	absDirPath := absPath
	if !isAbsPathADir {
		absDirPath = filepath.Dir(absPath)
	}

	if modulePackagePath == "" {
		modulePackagePath, err = gopkg.GetModulePackagePath(absDirPath)
		if err != nil {
			logrus.Errorf(" error: %s", err)
			return
		}
	}

	files = make([]*ast.TokenFile, 0)
	fileMap = make(map[string]*ast.TokenFile, 0)
	if len(fileNames) == 0 {
		return
	}

	for _, fileName := range fileNames {
		absFilePath := fmt.Sprintf("%s/%s", absDirPath, fileName)
		pos := &ast.FilePos{
			Package:  modulePackagePath,
			FileName: fileName,
			Name:     fileName,
			FilePath: absFilePath,
		}
		tokenFile := ast.NewTokenFile(pos)

		fileMap[absFilePath] = tokenFile
		files = append(files, tokenFile)
	}

	return
}

type Scanner struct {
	reader io.Reader
	file   *ast.TokenFile
	src    []byte // file src

	// scanning state
	// refer to go/scanner
	ch         rune // current character
	offset     int  // current character offset
	lineOffset int  // current line offset

	lineNo int // current line number

	rdOffset int // reading offset (position after current character)

	ErrorList ErrorList
}

func NewScanner(
	astFile *ast.TokenFile,
	reader io.Reader,
) (scanner *Scanner, err error) {
	var source []byte
	if reader != nil {
		source, err = ioutil.ReadAll(reader)
		if err != nil {
			logrus.Errorf("read src failed. error: %s", err)
			return
		}
	} else {
		source, err = ioutil.ReadFile(astFile.Pos.FilePath)
		if err != nil {
			logrus.Errorf("read file bytes failed. error: %s", err)
			return
		}
		reader = bytes.NewBuffer(source)
	}

	scanner = &Scanner{
		file:   astFile,
		reader: reader,
		src:    source,
	}
	err = scanner.init()
	if err != nil {
		logrus.Errorf("init scanner failed. error: %s", err)
		return
	}
	return
}

const (
	bom = 0xFEFF // byte order mark, only permitted as very first character
	eof = -1     // end of line char replacer
)

func (m *Scanner) init() (err error) {
	fileInfo, err := os.Stat(m.file.Pos.FilePath)
	if err != nil {
		logrus.Errorf("get file stat failed. path: %s, error: %s", m.file.Pos.FilePath, err)
		return
	}

	lenSrc := int64(len(m.src))

	if fileInfo.Size() != lenSrc {
		err = fmt.Errorf("file size (%d) does not match src len (%d)", fileInfo.Size(), lenSrc)
		return
	}

	// read first character
	m.next()
	if m.ch == bom { // byte order mark
		m.next()
	}

	return
}

func (m *Scanner) next() {
	if m.rdOffset >= len(m.src) { // end of file
		m.ch = eof
		m.offset = len(m.src)
		if m.ch == '\n' {
			m.lineOffset = m.offset
			m.file.AddLineOffset(m.offset)
		}
		return
	}

	// read a char
	m.offset = m.rdOffset
	if m.ch == '\n' { // new line position
		m.lineOffset = m.offset
		m.file.AddLineOffset(m.offset)
	}

	chRune := rune(m.src[m.rdOffset])
	chWidth := 1
	switch {
	case chRune == 0:
		m.error(m.offset, "illegal character NUL")
	case chRune >= utf8.RuneSelf: // not ASCII
		chRune, chWidth = utf8.DecodeRune(m.src[m.rdOffset:]) // decode an rune
		if chRune == utf8.RuneError && chWidth == 1 {
			m.error(m.offset, "illegal UTF-8 encoding")
		} else if chRune == bom && m.offset > 0 {
			m.error(m.offset, "illegal type order mark")
		}
	}

	m.rdOffset += chWidth
	m.ch = chRune
}

func (m *Scanner) error(offset int, msg string) {
	tokenPos := &ast.TokenPos{
		FilePos: m.file.Pos,
		Offset:  offset,
	}

	m.ErrorList.Add(tokenPos, msg)
}

// peek returns the byte following the most recently read character without
// advancing the scanner. If the scanner is at EOF, peek returns 0.
func (m *Scanner) peek() byte {
	if m.rdOffset < len(m.src) {
		return m.src[m.rdOffset]
	}
	return 0
}

func (m *Scanner) scan() {

}

func isLetter(ch rune) bool {
	return 'a' <= lower(ch) && lower(ch) <= 'z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return isDecimal(ch) || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

func lower(ch rune) rune     { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }
func isHex(ch rune) bool     { return '0' <= ch && ch <= '9' || 'a' <= lower(ch) && lower(ch) <= 'f' }

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= lower(ch) && lower(ch) <= 'f':
		return int(lower(ch) - 'a' + 10)
	}
	return 16 // larger than any legal digit val
}

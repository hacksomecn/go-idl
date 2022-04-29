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

package scanner

import (
	"fmt"
	"github.com/hacksomecn/go-idl/gopkg"
	"github.com/hacksomecn/go-idl/parser/ast"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"path/filepath"
)

// ScanFiles scan .gidl files in dir
func ScanFiles(
	path string,              // file or dir path
	modulePackagePath string, // package path with module name, if empty will find package for path
) (
	files []*ast.IdlFile,
	fileMap map[string]*ast.IdlFile, // file abs path -> *ast.IdlFile
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

	files = make([]*ast.IdlFile, 0)
	fileMap = make(map[string]*ast.IdlFile, 0)
	if len(fileNames) == 0 {
		return
	}

	for _, fileName := range fileNames {
		absFilePath := fmt.Sprintf("%s/%s", absDirPath, fileName)
		astFile := ast.NewIdlFile()
		astFile.Pos = &ast.FilePos{
			Package:  modulePackagePath,
			FileName: fileName,
			Name:     fileName,
			FilePath: absFilePath,
		}

		fileMap[absFilePath] = astFile
		files = append(files, astFile)
	}

	return
}

type Scanner struct {
	reader io.Reader
	file   *ast.IdlFile
	source []byte // file source

	// scanning state
	// refer to go/scanner
	ch         rune // current character
	offset     int  // character offset
	rdOffset   int  // reading offset (position after current character)
	lineOffset int  // current line offset
}

func NewScanner(
	astFile *ast.IdlFile,
	reader io.Reader,
) (scanner *Scanner, err error) {
	source, err := ioutil.ReadAll(reader)
	if err != nil {
		logrus.Errorf("read source failed. error: %s", err)
		return
	}

	scanner = &Scanner{
		file:   astFile,
		reader: reader,
		source: source,
	}
	return
}

func (m *Scanner) nextChar() {

}

func (m *Scanner) nextToken() {

}

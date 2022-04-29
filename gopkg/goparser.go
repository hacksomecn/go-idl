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

package gopkg

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"go/parser"
	"go/token"
	"path/filepath"
)

var fileset *token.FileSet

func init() {
	fileset = token.NewFileSet()
}

func parseDirPackageName(dir string) (packageName string, err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		logrus.Errorf("get abs path of dir %s failed. error: %s", dir, err)
		return
	}

	pkgs, err := parser.ParseDir(fileset, dir, nil, parser.PackageClauseOnly)
	if err != nil {
		logrus.Errorf("parse dir for finding package name failed. error: %s", err)
		return
	}

	pkgNames := make([]string, 0)
	for key, _ := range pkgs {
		pkgNames = append(pkgNames, key)
	}

	lenPkgNames := len(pkgNames)
	if lenPkgNames > 1 {
		err = fmt.Errorf("syntax error found: multiple package name in dir. dir: %s, names: %s", dir, pkgNames)
		return

	} else if lenPkgNames == 1 {
		packageName = pkgNames[0]

	} else {
		packageName = filepath.Base(dir)
	}

	return
}

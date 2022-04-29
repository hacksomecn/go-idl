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
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

const FilePostfix = ".gidl"

func IsIdlFile(filename string) (yes bool) {
	return filepath.Ext(filename) == FilePostfix
}

func FindIdlFiles(path string) (fileNames []string, err error) {
	fileNames = make([]string, 0)

	pathInfo, err := os.Stat(path)
	if err != nil {
		logrus.Errorf("no such file or directory. error: %s", err)
		return
	}

	if pathInfo.IsDir() {
		dirFiles, errR := os.ReadDir(path)
		err = errR
		if err != nil {
			logrus.Errorf("read dir failed. path: %s, error: %s", path, err)
			return
		}

		for _, file := range dirFiles {
			if IsIdlFile(file.Name()) {
				fileNames = append(fileNames, file.Name())
			}
		}

	} else if IsIdlFile(path) {
		_, fileName := filepath.Split(path)
		fileNames = append(fileNames, fileName)
	}

	return
}

func PathExists(
	path string,
) (
	exists bool,
	isDir bool,
) {
	info, err := os.Stat(path)
	if nil != err {
		return
	}

	exists = true
	isDir = info.IsDir()
	return
}

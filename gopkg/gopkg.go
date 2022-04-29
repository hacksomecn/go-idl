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
	"encoding/json"
	"fmt"
	"github.com/hacksomecn/go-idl/syspkg"
	"github.com/sirupsen/logrus"
	"golang.org/x/mod/modfile"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	GoPath        string        // go path
	GoRoot        string        // go root
	GoExecutable  string        // go binary program
	GoEnvs        GoEnvParams   // go environment
	GoModFile     *modfile.File // project go mod file object
	GoModFilePath string        // go mod file path
	GoModDir      string        // directory where go.mod placed
	GoModName     string        // project go module name
)

type GoEnvParams struct { // go env -json
	GoMod      string `json:"GOMOD"`
	GoModCache string `json:"GOMODCACHE"`
}

func init() {
	err := loadEnv()
	if err != nil {
		logrus.Panicf("get go env value failed. %s", err)
		return
	}
}

func loadEnv() (err error) {
	GoRoot = os.Getenv("GOROOT")
	if GoRoot == "" {
		err = fmt.Errorf("can not find GOROOT")
		logrus.Error(err)
		return
	}

	GoPath = os.Getenv("GOPATH")
	GoExecutable = fmt.Sprintf("%s/bin/go", GoRoot)

	exit, output, err := syspkg.RunCommand("", GoExecutable, "env", "-json")
	if err != nil {
		logrus.Errorf("run go env -json failed. error: %s", err)
		return
	}
	if exit != 0 {
		err = fmt.Errorf("exec `go env -json` failed. error: %s", err)
		logrus.Error(err)
		return
	}

	err = json.Unmarshal([]byte(output), &GoEnvs)
	if err != nil {
		logrus.Errorf("unmarshal go env json string to struct failed. value: %s, error: %s", output, err)
		return
	}

	if GoEnvs.GoMod == "" {
		err = fmt.Errorf("can not find GOMOD from `go env -json`")
		logrus.Error(err)
		return
	}

	GoModFilePath = GoEnvs.GoMod
	GoModFilePath, err = filepath.Abs(GoModFilePath)
	if err != nil {
		logrus.Errorf("get abs mod file path failed. error: %s", err)
		return
	}

	GoModFile, err = readGoModFile(GoModFilePath)
	if err != nil {
		logrus.Errorf("get go mod faile failed. %s", err)
		return
	}

	GoModDir = filepath.Dir(GoEnvs.GoMod)
	GoModName = GoModFile.Module.Mod.Path
	return
}

func readGoModFile(goModFilePath string) (goMod *modfile.File, err error) {
	data, err := ioutil.ReadFile(goModFilePath)
	if err != nil {
		logrus.Errorf("read go mod failed. path: %s, error: %s", goModFilePath, err)
		return
	}

	goMod, err = modfile.Parse(goModFilePath, data, nil)
	if err != nil {
		logrus.Errorf("parse go mod file fialed. path: %s, error: %s", goModFilePath, err)
		return
	}

	return
}

// GetModulePackagePath find dir package path
func GetModulePackagePath(
	dirPath string,
) (
	packagePath string, // package path under module
	err error,
) {
	dirPath, err = filepath.Abs(dirPath)
	if err != nil {
		logrus.Errorf("get abs path failed. path: %s, error: %s", dirPath, err)
		return
	}

	if !strings.HasPrefix(dirPath, GoModDir) {
		err = fmt.Errorf("dir %s should be under go module: %s", dirPath, GoModDir)
		return
	}

	relPath := strings.Replace(dirPath, GoModDir, "", 1)
	relPathBase := filepath.Base(dirPath)
	dirPackageName, err := parseDirPackageName(dirPath)
	if err != nil {
		logrus.Errorf("parse dir package name failed. error: %s", err)
		return
	}

	relPath = strings.Replace(relPath, relPathBase, dirPackageName, 1)
	packagePath = fmt.Sprintf("%s%s", GoModName, relPath)

	return
}

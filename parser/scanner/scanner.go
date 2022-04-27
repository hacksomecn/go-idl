package scanner

import (
	"fmt"
	"github.com/hacksomecn/go-idl/gopkg"
	"github.com/hacksomecn/go-idl/parser/ast"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

func Scan(
	dirPath string, // file dir path
	modulePackagePath string, // package path with module name, if empty will find package for path
) (files []*ast.IdlFile, err error) {
	// get package name
	fileNames, err := FindIdlFiles(dirPath)
	if err != nil {
		logrus.Errorf("get idl files failed. error: %s", err)
		return
	}

	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		logrus.Errorf("get abs path failed. path: %s, error: %s", dirPath, err)
		return
	}

	if modulePackagePath == "" {
		modulePackagePath, err = gopkg.GetModulePackagePath(dirPath)
		if err != nil {
			logrus.Errorf(" error: %s", err)
			return
		}
	}

	files = make([]*ast.IdlFile, 0)
	if len(fileNames) == 0 {
		return
	}

	for _, fileName := range fileNames {
		absFilePath := fmt.Sprintf("%s/%s", absDirPath, fileName)
		astFile := ast.NewIdlFile()
		astFile.Pos = &ast.Pos{
			Package:  modulePackagePath,
			FileName: fileName,
			Name:     fileName,
			FilePath: absFilePath,
		}

		files = append(files, astFile)
	}

	return
}

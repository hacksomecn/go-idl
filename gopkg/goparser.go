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

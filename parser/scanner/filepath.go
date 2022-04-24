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

func FindIdlFiles(path string) (files []string, err error) {
	files = make([]string, 0)

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
				files = append(files, file.Name())
			}
		}

	} else if IsIdlFile(path) {
		files = append(files, path)
	}

	return
}

func PathExists(path string) (exists bool) {
	_, err := os.Stat(path)
	if nil != err {
		return
	}

	exists = true
	return
}

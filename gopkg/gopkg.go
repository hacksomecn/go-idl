package gopkg

import (
	"encoding/json"
	"fmt"
	"github.com/hacksomecn/go-idl/syspkg"
	"github.com/sirupsen/logrus"
	"golang.org/x/mod/modfile"
	"io/ioutil"
	"os"
)

var (
	GoPath       string
	GoRoot       string
	GoExecutable string
	GoEnvs       GoEnvParams
	GoModFile    *modfile.File
)

type GoEnvParams struct { // go env -json
	GoMod      string `json:"GOMOD"`
	GoModCache string `json:"GOMODCACHE"`
}

func GoModName() string {
	return GoModFile.Module.Mod.Path
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

	GoModFile, err = readGoModFile(GoEnvs.GoMod)
	if err != nil {
		logrus.Errorf("get go mod faile failed. %s", err)
		return
	}
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

func FindModulePackagePath(
	dirPath string,
) (
	packagePath string, // package path under module
	err error,
) {
	// TODO
	err = fmt.Errorf("need implement FindModulePackagePath")
	return
}

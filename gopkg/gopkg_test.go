package gopkg

import (
	"fmt"
	"testing"
)

func TestGoModFile(t *testing.T) {
	fmt.Println(GoModFile.Module.Mod)
	fmt.Println(GoModFilePath)
	fmt.Println(GoModDir)
	fmt.Println(GoModName)
}

func TestGetModulePackagePath(t *testing.T) {
	data := []struct {
		Dir         string
		PackagePath string
	}{
		{
			Dir:         "/Users/hao/Documents/Projects/Github/go-idl/example/files/dir1",
			PackagePath: "github.com/hacksomecn/go-idl/example/files/dir1",
		},
		{
			Dir:         "/Users/hao/Documents/Projects/Github/go-idl/example/files/nogodir",
			PackagePath: "github.com/hacksomecn/go-idl/example/files/nogodir",
		},
		{
			Dir:         "/Users/hao/Documents/Projects/Github/go-idl/example/files",
			PackagePath: "github.com/hacksomecn/go-idl/example/filess",
		},
	}
	for _, item := range data {
		t.Run(item.Dir, func(t *testing.T) {
			packagePath, err := GetModulePackagePath(item.Dir)
			if err != nil {
				t.Error(err)
				return
			}
			if packagePath != item.PackagePath {
				t.Errorf("not equal %s %s", packagePath, item.PackagePath)
			}

			fmt.Println(packagePath)
		})
	}
}

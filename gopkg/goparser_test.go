package gopkg

import (
	"fmt"
	"testing"
)

func TestParseDirPackage(t *testing.T) {
	packageName, err := parseDirPackageName("/Users/hao/Documents/Projects/Github/go-idl/example/files/nogodir")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(packageName)
}

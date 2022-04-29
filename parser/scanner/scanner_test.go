package scanner

import (
	"fmt"
	"testing"
)

func TestFindIdlFiles(t *testing.T) {
	path := "/Users/hao/Documents/Projects/Github/go-idl/example/files"
	files, _, err := ScanFiles(path, "")
	if err != nil {
		t.Error(err)
		return
	}

	for _, file := range files {
		fmt.Printf("%+v\n", file.Pos)
	}
}

func TestScanFiles(t *testing.T) {
	path := "/Users/hao/Documents/Projects/Github/go-idl/example/decl/decl.gidl"
	files, _, err := ScanFiles(path, "")
	if err != nil {
		t.Error(err)
		return
	}

	_ = files
}

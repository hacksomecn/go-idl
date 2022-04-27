package scanner

import (
	"fmt"
	"testing"
)

func TestFindIdlFiles(t *testing.T) {
	path := "/Users/hao/Documents/Projects/Github/go-idl/example/files"
	files, err := Scan(path, "")
	if err != nil {
		t.Error(err)
		return
	}

	for _, file := range files {
		fmt.Printf("%+v", file.Pos)
	}
}

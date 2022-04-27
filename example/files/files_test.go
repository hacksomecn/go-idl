package filess

import (
	"fmt"
	"github.com/hacksomecn/go-idl/parser/scanner"
	"testing"
)

func TestScan(t *testing.T) {
	files, err := scanner.Scan("./dir1", "")
	if err != nil {
		t.Error(err)
		return
	}

	for _, file := range files {
		fmt.Println(file.Pos)
	}
}

func TestName(t *testing.T) {
	fmt.Println("2")
}

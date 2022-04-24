package files

import (
	"fmt"
	"github.com/hacksomecn/go-idl/parser/scanner"
	"testing"
)

func TestScan(t *testing.T) {
	files, err := scanner.Scan("./")
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(files)
}

func TestName(t *testing.T) {
	fmt.Println("2")
}

package scanner

import (
	"fmt"
	"github.com/hacksomecn/go-idl/parser/ast"
	"io/ioutil"
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

func TestNext(t *testing.T) {
	path := "/Users/hao/Documents/Projects/Github/go-idl/example/idlfile/go-idl.gidl"
	files, _, err := ScanFiles(path, "")
	if err != nil {
		t.Error(err)
		return
	}
	file := files[0]

	src, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
		return
	}

	scanner, err := NewScanner(file, src)
	if err != nil {
		t.Error(err)
		return
	}

	for {
		if scanner.ch == eof {
			break
		}

		fmt.Printf("%s", string(scanner.ch))
		scanner.next()
	}
}

func TestScan(t *testing.T) {
	path := "/Users/hao/Documents/Projects/Github/go-idl/example/idlfile/go-idl.gidl"
	files, _, err := ScanFiles(path, "")
	if err != nil {
		t.Error(err)
		return
	}
	file := files[0]

	src, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
		return
	}

	scanner, err := NewScanner(file, src)
	if err != nil {
		t.Error(err)
		return
	}

	for {
		pos, token, lit := scanner.Scan()
		fmt.Println(pos, token, lit)
		if token == ast.EOF {
			break
		}
	}
}

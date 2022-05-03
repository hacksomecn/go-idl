/*
 * The MIT License (MIT)
 *
 * Copyright © 2022 Hao Luo <haozzzzzzzz@gmail.com>

 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package parser

import (
	"fmt"
	"github.com/hacksomecn/go-idl/parser/ast"
	"github.com/hacksomecn/go-idl/parser/scanner"
	"reflect"
	"testing"
)

func TestParseFile(t *testing.T) {
	files, _, err := scanner.ScanFiles("/Users/hao/Documents/Projects/Github/go-idl/example/idlfile/go-idl.gidl", "")
	if err != nil {
		t.Error(err)
		return
	}
	file := files[0]
	parser, err := NewParser(file)
	if err != nil {
		t.Error(err)
		return
	}

	idlFile := parser.parseFile()
	for _, del := range idlFile.Imports {
		fmt.Println(string(del.Expr))
	}
}

func TestParseFileAssign(t *testing.T) {
	files, _, err := scanner.ScanFiles("/Users/hao/Documents/Projects/Github/go-idl/example/idlfile/go-idl.gidl", "")
	if err != nil {
		t.Error(err)
		return
	}
	file := files[0]
	parser, err := NewParser(file)
	if err != nil {
		t.Error(err)
		return
	}

	idlFile := parser.parseFile()
	for _, del := range idlFile.Assigns {
		fmt.Println(string(del.Expr))
		fmt.Println(del.Tok)
		fmt.Println(del.Spec.Name, del.Spec.Value)
	}
}

func TestPrintToken0(t *testing.T) {
	files, _, err := scanner.ScanFiles("/Users/hao/Documents/Projects/Github/go-idl/example/idlfile/model.gidl", "")
	if err != nil {
		t.Error(err)
		return
	}

	file := files[0]
	parser, err := NewParser(file)
	if err != nil {
		t.Error(err)
		return
	}

	parser.printToken0()
	err = parser.errors.Err()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestParseFileModel(t *testing.T) {
	files, _, err := scanner.ScanFiles("/Users/hao/Documents/Projects/Github/go-idl/example/idlfile/model.gidl", "")
	if err != nil {
		t.Error(err)
		return
	}

	file := files[0]
	parser, err := NewParser(file)
	if err != nil {
		t.Error(err)
		return
	}

	idlFile := parser.parseFile()
	err = parser.errors.Err()
	if err != nil {
		t.Error(err)
		return
	}

	for idx, model := range idlFile.Models {
		_ = idx
		fmt.Println(model.Pos)
		fmt.Println(string(model.Expr))

		fmt.Println("model doc:", model.Doc)
		fmt.Println("model comment:", model.Comment)
		fmt.Println("model token:", model.Tok)
		fmt.Println("model spec:")
		spec := model.Spec
		fmt.Println("\tpos:", spec.TypePos)
		fmt.Println("\tdoc: ", spec.Doc)
		fmt.Println("\tcomment: ", spec.Comment)
		fmt.Println("\tname: ", spec.Name)
		fmt.Println("\tfields:")
		for _, field := range spec.Fields {
			fmt.Println("\t\tpos:", field.Pos)
			fmt.Println("\t\tdoc:", field.Doc)
			fmt.Println("\t\tcomment: ", field.Comment)
			fmt.Println("\t\tname: ", field.Name)
			if field.Tag != nil {
				fmt.Printf("\t\ttag: %+v\n", field.Tag.Value)
			}
			fmt.Println("\t\texported:", field.Exported)
			fmt.Println("\t\tembedded:", field.Embedded)
			fmt.Println("\t\ttype:", field.Type.TypeNameIdent().Name, field.Type)
			switch subType := field.Type.(type) {
			case *ast.ModelType:
				for _, ff := range subType.Fields {
					fmt.Println("\t\t\tpos:", ff.Pos)
					fmt.Println("\t\t\tdoc:", ff.Doc)
					fmt.Println("\t\t\tcomment:", ff.Comment)
					fmt.Println("\t\t\tname:", ff.Name)
					if ff.Tag != nil {
						fmt.Printf("\t\t\ttag: %+v\n", ff.Tag.Value)
					}
					fmt.Println("\t\t\texported", ff.Exported)
					fmt.Println("\t\t\tembedded", ff.Embedded)
					fmt.Println("\t\t\ttype:", ff.Type)
					switch ttt := ff.Type.(type) {
					case *ast.ModelType:
						for _, fff := range ttt.Fields {
							fmt.Println("\t\t\t\tpos:", fff.Pos)
							fmt.Println("\t\t\t\tdoc:", fff.Doc)
							fmt.Println("\t\t\t\tcomment:", fff.Comment)
							fmt.Println("\t\t\t\tname:", fff.Name)
							if fff.Tag != nil {
								fmt.Printf("\t\t\t\ttag: %+v\n", fff.Tag.Value)
							}
							fmt.Println("\t\t\t\texported", fff.Exported)
							fmt.Println("\t\t\t\tembedded", fff.Embedded)
							fmt.Println("\t\t\t\ttype:", fff.Type)
							fmt.Println()
						}
					}
					fmt.Println()
				}
			case *ast.ArrayType:
				fmt.Println("\t\t\telem:", reflect.TypeOf(subType.ElemType), subType.ElemType.TypeNameIdent())
			case *ast.MapType:
				fmt.Println("\t\t\tkey:", reflect.TypeOf(subType.KeyType), subType.KeyType.TypeNameIdent())
				fmt.Println("\t\t\telement:", reflect.TypeOf(subType.ElemType), subType.ElemType.TypeNameIdent())
			}
			fmt.Println()
		}
	}
}

package goq_test

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	. "github.com/tenntenn/goq"
	"github.com/tenntenn/optional"
)

func TestExec(t *testing.T) {
	const src = `package main
	func main() {
		n := 10
		println(n)
	}`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", src, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	config := &types.Config{
		Importer: importer.Default(),
	}

	info := &types.Info{
		Defs: map[*ast.Ident]types.Object{},
		Uses: map[*ast.Ident]types.Object{},
	}

	if _, err := config.Check("main", fset, []*ast.File{f}, info); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q := &Int{}
	results := New(fset, []*ast.File{f}, info).Query(q)
	if len(results) != 2 {
		t.Fatalf("the number of result must be 2 but %d", len(results))
	}

	if n := results[0].Object.Name(); n != "n" {
		t.Errorf(`exepect object name is "n" but %q`, n)
	}
}

func TestError(t *testing.T) {
	const src = `package main
	type Err string
	func (err Err) Error() string {return string(err)}
	func f() (int, Err) { // 4
		return 0, Err("hoge")
	}
	func main() {
		println(f()) // 8
		println(func() Err { // 9
			return Err("fuga")
		}())
		println(func() error { // 12
			return nil
		}())
	}`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", src, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	config := &types.Config{
		Importer: importer.Default(),
	}

	info := &types.Info{
		Defs:  map[*ast.Ident]types.Object{},
		Uses:  map[*ast.Ident]types.Object{},
		Types: map[ast.Expr]types.TypeAndValue{},
	}

	if _, err := config.Check("main", fset, []*ast.File{f}, info); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errorType := types.Universe.Lookup("error").Type()
	results := New(fset, []*ast.File{f}, info).Query(&Signature{
		Results: optional.NewTupple(nil, func(v1, v2 interface{}) bool {
			q, ok := v1.(Query)
			if !ok {
				return false
			}

			v, ok := v2.(*types.Var)
			if !ok {
				return false
			}

			return q.Match(v)
		}).Put(-1, &Var{
			Type: And(&Type{
				Implements: errorType,
			}, Not(&Type{
				Identical: errorType,
			})),
		}),
	})

	/*
		if len(results) != 2 {
			t.Fatalf("the number of result must be 2 but %d", len(results))
		}
	*/

	for _, r := range results {
		fmt.Println(fset.Position(r.Node.Pos()))
	}
	/*
		if n := results[0].Object.Name(); n != "n" {
			t.Errorf(`exepect object name is "n" but %q`, n)
		}
	*/
}

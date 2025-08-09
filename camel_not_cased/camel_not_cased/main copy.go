package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"regexp"
	"strings"
)

// camelToSnake converts camelCase or PascalCase to snake_case
func camelToSnake(s string) string {
	if s == "" || s == "_" || len(s) == 1 {
		return s
	}
	re1 := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	s = re1.ReplaceAllString(s, `${1}_${2}`)
	re2 := regexp.MustCompile(`([A-Z])([A-Z][a-z])`)
	s = re2.ReplaceAllString(s, `${1}_${2}`)
	return strings.ToLower(s)
}

// isSelectorSel checks if the given ident is the .Sel part of a SelectorExpr (e.g., x.Sel)
func isSelectorSel(id *ast.Ident, parents []ast.Node) bool {
	if len(parents) == 0 {
		return false
	}
	// The direct parent might be a SelectorExpr whose Sel is this ident
	if sel, ok := parents[len(parents)-1].(*ast.SelectorExpr); ok && sel.Sel == id {
		return true
	}
	return false
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: camelnotcased <file.go>\n")
		os.Exit(2)
	}
	filename := os.Args[1]

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse: %v\n", err)
		os.Exit(1)
	}

	renames := map[*ast.Object]string{}

	var parents []ast.Node

	// First pass: decide what to rename
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			if len(parents) > 0 {
				parents = parents[:len(parents)-1]
			}
			return true
		}
		parents = append(parents, n)

		if id, ok := n.(*ast.Ident); ok {
			if id.Obj == nil || isSelectorSel(id, parents) {
				return true
			}
			// Skip exported names, package names, and types
			if id.IsExported() || id.Obj.Kind == ast.Pkg || id.Obj.Kind == ast.Typ {
				return true
			}
			// Only rename local vars, params, receivers, results
			switch id.Obj.Kind {
			case ast.Var, ast.Con:
				old := id.Name
				newName := camelToSnake(old)
				if newName != old && isValidIdent(newName) {
					if _, done := renames[id.Obj]; !done {
						renames[id.Obj] = newName
					}
				}
			}
		}
		return true
	})

	// Second pass: apply renames
	parents = nil
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			if len(parents) > 0 {
				parents = parents[:len(parents)-1]
			}
			return true
		}
		parents = append(parents, n)

		if id, ok := n.(*ast.Ident); ok {
			if id.Obj == nil || isSelectorSel(id, parents) {
				return true
			}
			if newName, ok := renames[id.Obj]; ok {
				id.Name = newName
			}
		}
		return true
	})

	var out bytes.Buffer
	cfg := &printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}
	if err := cfg.Fprint(&out, fset, file); err != nil {
		fmt.Fprintf(os.Stderr, "print: %v\n", err)
		os.Exit(1)
	}
	os.Stdout.Write(out.Bytes())
}

func isValidIdent(s string) bool {
	if s == "" {
		return false
	}
	if s[0] != '_' && (s[0] < 'a' || s[0] > 'z') {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if c == '_' || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			continue
		}
		return false
	}
	return true
}
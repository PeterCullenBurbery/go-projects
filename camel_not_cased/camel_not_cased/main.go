package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/types"
	"os"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	keepExported = flag.Bool("keep-exported", false, "do not rename exported identifiers")
)

// camelToSnake converts camelCase/PascalCase to snake_case
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

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: camelnotcased [flags] <packages-or-files>\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	// Load with full type info
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes,
	}

	pkgs, err := packages.Load(cfg, flag.Args()...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	for _, pkg := range pkgs {
		renamePkg(pkg)
	}
}

func renamePkg(pkg *packages.Package) {
	fset := pkg.Fset
	info := pkg.TypesInfo
	thisPkg := pkg.Types

	// Decide new names: map each *types.Object -> new name
	renames := map[types.Object]string{}

	// Helper to consider an object for rename
	consider := func(obj types.Object) {
		if obj == nil {
			return
		}
		// Only objects that belong to *this* package
		if obj.Pkg() == nil || obj.Pkg() != thisPkg {
			return
		}
		// Optionally skip exported
		if *keepExported && obj.Exported() {
			return
		}
		// Skip package names
		if _, isPkgName := obj.(*types.PkgName); isPkgName {
			return
		}
		// Compute new name
		old := obj.Name()
		newName := camelToSnake(old)
		if newName == old || !isValidIdent(newName) {
			return
		}
		renames[obj] = newName
	}

	// 1) Collect renames by scanning declarations
	for ident, obj := range info.Defs {
		if ident == nil || obj == nil {
			continue
		}
		// Skip blank identifiers
		if ident.Name == "_" {
			continue
		}
		consider(obj)
	}
	// Also consider implicit objects that appear only in Uses (e.g., embedded method sets not explicitly named)
	for _, obj := range info.Uses {
		consider(obj)
	}
	// Also look at Selections for methods/fields reached via selector
	for _, sel := range info.Selections {
		if sel == nil {
			continue
		}
		consider(sel.Obj())
	}

	// Nothing to do?
	if len(renames) == 0 {
		return
	}

	// 2) Apply renames across all syntax trees
	for i, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.Ident:
				if n.Name == "_" {
					return true
				}
				if obj := info.ObjectOf(n); obj != nil {
					if newName, ok := renames[obj]; ok {
						n.Name = newName
					}
				}
			case *ast.SelectorExpr:
				// Selections carry type info; only rename when the selected object is ours
				if sel := info.Selections[n]; sel != nil {
					if newName, ok := renames[sel.Obj()]; ok {
						n.Sel.Name = newName
					}
				}
			}
			return true
		})

		// 3) Print updated file to stdout (one after another)
		var out bytes.Buffer
		cfg := &printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}
		if err := cfg.Fprint(&out, fset, file); err != nil {
			fmt.Fprintf(os.Stderr, "print %s: %v\n", pkg.CompiledGoFiles[i], err)
			os.Exit(1)
		}
		os.Stdout.Write(out.Bytes())
		fmt.Println() // separator newline between files
	}
}

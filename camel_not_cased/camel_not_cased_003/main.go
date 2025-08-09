package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	keepExported = flag.Bool("keep-exported", false, "do not rename exported identifiers")
	writeFiles   = flag.Bool("w", false, "write the results to the source files instead of stdout")
	editComments = flag.Bool("comments", true, "also rewrite comments that mention renamed identifiers")
	editStrings  = flag.Bool("strings", false, "also rewrite string literals that mention renamed identifiers (careful)")
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
		if err := renamePkg(pkg); err != nil {
			fmt.Fprintf(os.Stderr, "rename: %v\n", err)
			os.Exit(1)
		}
	}
}

func renamePkg(pkg *packages.Package) error {
	fset := pkg.Fset
	info := pkg.TypesInfo
	thisPkg := pkg.Types

	// Decide new names: map each *types.Object -> new name
	renames := map[types.Object]string{}
	// Also keep a simple old->new text map for comments/strings
	oldToNew := map[string]string{}

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
		old := obj.Name()
		newName := camelToSnake(old)
		if newName == old || !isValidIdent(newName) {
			return
		}
		renames[obj] = newName
		oldToNew[old] = newName
	}

	// Collect candidates from Defs (declarations)
	for ident, obj := range info.Defs {
		if ident == nil || obj == nil {
			continue
		}
		if ident.Name == "_" {
			continue
		}
		consider(obj)
	}
	// Also consider objects that appear only in Uses
	for _, obj := range info.Uses {
		consider(obj)
	}
	// And the selected objects from selectors (fields/methods)
	for _, sel := range info.Selections {
		if sel == nil {
			continue
		}
		consider(sel.Obj())
	}

	// Nothing to do?
	if len(renames) == 0 {
		// Still print or write original files for consistency? We'll just no-op.
		if !*writeFiles {
			for i, file := range pkg.Syntax {
				var out bytes.Buffer
				cfg := &printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}
				if err := cfg.Fprint(&out, fset, file); err != nil {
					return fmt.Errorf("print %s: %w", pkg.CompiledGoFiles[i], err)
				}
				os.Stdout.Write(out.Bytes())
				fmt.Println()
			}
		}
		return nil
	}

	// Prebuild word-boundary regexes for comment/string rewriting
	var wordRegex map[string]*regexp.Regexp
	if *editComments || *editStrings {
		wordRegex = make(map[string]*regexp.Regexp, len(oldToNew))
		for old := range oldToNew {
			wordRegex[old] = regexp.MustCompile(`\b` + regexp.QuoteMeta(old) + `\b`)
		}
	}

	for i, file := range pkg.Syntax {
		// 1) Rename identifiers & selector .Sel where the selected object is ours
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
				if sel := info.Selections[n]; sel != nil {
					if newName, ok := renames[sel.Obj()]; ok {
						n.Sel.Name = newName
					}
				}
			}
			return true
		})

		// 2) Optionally rewrite comments
		if *editComments {
			rewriteComments(file, oldToNew, wordRegex)
		}

		// 3) Optionally rewrite string literals
		if *editStrings {
			rewriteStrings(file, oldToNew, wordRegex)
		}

		// 4) Output
		var out bytes.Buffer
		cfg := &printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}
		if err := cfg.Fprint(&out, fset, file); err != nil {
			return fmt.Errorf("print %s: %w", pkg.CompiledGoFiles[i], err)
		}

		if *writeFiles {
			if err := os.WriteFile(pkg.CompiledGoFiles[i], out.Bytes(), 0o666); err != nil {
				return fmt.Errorf("write %s: %w", pkg.CompiledGoFiles[i], err)
			}
		} else {
			os.Stdout.Write(out.Bytes())
			fmt.Println()
		}
	}

	return nil
}

func rewriteComments(file *ast.File, oldToNew map[string]string, regs map[string]*regexp.Regexp) {
	if file == nil || file.Comments == nil || len(oldToNew) == 0 {
		return
	}
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			text := c.Text
			for old, re := range regs {
				text = re.ReplaceAllString(text, oldToNew[old])
			}
			c.Text = text
		}
	}
}

func rewriteStrings(file *ast.File, oldToNew map[string]string, regs map[string]*regexp.Regexp) {
	if file == nil || len(oldToNew) == 0 {
		return
	}
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		val := lit.Value
		if len(val) < 2 {
			return true
		}
		quote := val[0]
		switch quote {
		case '"':
			// interpreted string
			unquoted, err := strconv.Unquote(val)
			if err != nil {
				return true
			}
			for old, re := range regs {
				unquoted = re.ReplaceAllString(unquoted, oldToNew[old])
			}
			lit.Value = strconv.Quote(unquoted)
		case '`':
			// raw string
			content := val[1 : len(val)-1]
			for old, re := range regs {
				content = re.ReplaceAllString(content, oldToNew[old])
			}
			// Caution: if content contains backticks, we could break; assume rare for identifier words.
			lit.Value = "`" + content + "`"
		default:
			// unknown quote style, skip
		}
		return true
	})
}
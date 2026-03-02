package osexit

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "osexit",
	Doc:      "forbid os.Exit, panic and log.Fatal outside main function",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	var inMainStack []bool
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
		(*ast.CallExpr)(nil),
	}

	insp.Nodes(nodeFilter, func(node ast.Node, push bool) bool {
		if push {
			switch node := node.(type) {
			case *ast.FuncDecl:
				inMainStack = append(inMainStack, pass.Pkg.Name() == "main" && node.Name.Name == "main")
			case *ast.FuncLit:
				inheritsMain := len(inMainStack) > 0 && inMainStack[len(inMainStack)-1]
				inMainStack = append(inMainStack, inheritsMain)
			case *ast.CallExpr:
				insideMain := containsTrue(inMainStack)
				if insideMain {
					return true
				}

				filename := pass.Fset.Position(node.Pos()).Filename
				if strings.Contains(filename, "go-build") || strings.Contains(filename, "Library/Caches") {
					return true
				}

				if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
						pass.Reportf(node.Pos(), "os.Exit must not be used outside main")
						return true
					}
				}

				if ident, ok := node.Fun.(*ast.Ident); ok && ident.Name == "panic" {
					pass.Reportf(node.Pos(), "panic must not be used outside main")
					return true
				}

				if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "log" {
						switch sel.Sel.Name {
						case "Fatal", "Fatalf", "Fatalln":
							pass.Reportf(node.Pos(), "log.%s must not be used outside main", sel.Sel.Name)
						}
					}
				}
			}
			return true
		}
		if _, ok := node.(*ast.FuncDecl); ok {
			inMainStack = inMainStack[:len(inMainStack)-1]
		}
		if _, ok := node.(*ast.FuncLit); ok {
			inMainStack = inMainStack[:len(inMainStack)-1]
		}
		return true
	})

	return nil, nil
}

func containsTrue(s []bool) bool {
	for _, x := range s {
		if x {
			return true
		}
	}
	return false
}

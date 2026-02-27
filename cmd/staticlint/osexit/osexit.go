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
	Doc:      "forbid direct os.Exit calls in main function of main package",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	var inMainStack []bool
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
		(*ast.CallExpr)(nil),
	}

	insp.Nodes(nodeFilter, func(n ast.Node, push bool) bool {
		if push {
			switch node := n.(type) {
			case *ast.FuncDecl:
				inMainStack = append(inMainStack, node.Name.Name == "main")
			case *ast.FuncLit:
				inMainStack = append(inMainStack, false)
			case *ast.CallExpr:
				if len(inMainStack) > 0 && inMainStack[len(inMainStack)-1] {
					filename := pass.Fset.Position(node.Pos()).Filename
					if strings.Contains(filename, "go-build") || strings.Contains(filename, "Library/Caches") {
						return true
					}
					if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
							pass.Reportf(node.Pos(), "direct os.Exit call in main is forbidden, use return or log.Fatal instead")
						}
					}
				}
			}
			return true
		}
		if _, ok := n.(*ast.FuncDecl); ok {
			inMainStack = inMainStack[:len(inMainStack)-1]
		}
		if _, ok := n.(*ast.FuncLit); ok {
			inMainStack = inMainStack[:len(inMainStack)-1]
		}
		return true
	})

	return nil, nil
}

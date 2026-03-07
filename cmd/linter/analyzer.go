package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "nolintpanic",
	Doc:  "checks for panic usage and forbidden log.Fatal/os.Exit calls",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {

		pkgName := pass.Pkg.Name()

		ast.Inspect(file, func(n ast.Node) bool {

			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// --- panic()
			if ident, ok := call.Fun.(*ast.Ident); ok {
				if ident.Name == "panic" {
					pass.Reportf(call.Pos(), "panic call is forbidden")
				}
			}

			// --- log.Fatal / os.Exit
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {

				if ident, ok := sel.X.(*ast.Ident); ok {

					if ident.Name == "log" && sel.Sel.Name == "Fatal" {
						if pkgName != "main" {
							pass.Reportf(call.Pos(), "log.Fatal is forbidden outside main")
						}
					}

					if ident.Name == "os" && sel.Sel.Name == "Exit" {
						if pkgName != "main" {
							pass.Reportf(call.Pos(), "os.Exit is forbidden outside main")
						}
					}

				}
			}

			return true
		})
	}

	return nil, nil
}

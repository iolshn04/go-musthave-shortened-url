package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "nolintpanic",
	Doc:  "checks for panic usage and forbidden log.Fatal/os.Exit calls",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	pkgName := pass.Pkg.Name()

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {

			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// panic()
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
				pass.Reportf(call.Pos(), "panic call is forbidden")
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}

			obj := pass.TypesInfo.Uses[ident]
			pkgNameObj, ok := obj.(*types.PkgName)
			if !ok {
				return true
			}

			importedPath := pkgNameObj.Imported().Path()

			if pkgName != "main" && importedPath == "log" && sel.Sel.Name == "Fatal" {
				pass.Reportf(call.Pos(), "log.Fatal is forbidden outside main")
			}

			if pkgName != "main" && importedPath == "os" && sel.Sel.Name == "Exit" {
				pass.Reportf(call.Pos(), "os.Exit is forbidden outside main")
			}

			return true
		})
	}

	return nil, nil
}

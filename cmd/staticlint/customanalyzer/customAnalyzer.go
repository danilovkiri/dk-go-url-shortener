// Package customanalyzer provides custom code analysis.
package customanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// OsExitInMainAnalyzer is the *analysis.Analyzer type for using in multichecker.
var OsExitInMainAnalyzer = &analysis.Analyzer{
	Name: "osexitmain",
	Doc:  "check for os.Exit call in main() body",
	Run:  run,
}

// run preforms analysis of code for presence of os.Exit() calls inside main() body.
func run(pass *analysis.Pass) (interface{}, error) {
	// iterate over input .go files
	for _, file := range pass.Files {
		// iterate over all AST nodes with ast.Inspect
		ast.Inspect(file, func(n ast.Node) bool {
			// look for `main` function declaration
			if v, ok := n.(*ast.FuncDecl); ok && v.Name.Name == `main` {
				// iterate over AST nodes in declaration body
				for _, stmt := range v.Body.List {
					// look for expression statements
					if ex, ok := stmt.(*ast.ExprStmt); ok {
						// look for function call expressions
						if call, ok := ex.X.(*ast.CallExpr); ok {
							// look for selector expressions
							if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
								// check that selector's X.Name is `os` and Sel.Name is `Exit`
								if i, ok := selector.X.(*ast.Ident); ok && i.Name == `os` {
									if selector.Sel.Name == `Exit` {
										pass.Reportf(selector.Pos(), "call to os.Exit in main body")
									}
								}
							}
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}

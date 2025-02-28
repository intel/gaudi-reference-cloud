package disallow_int_formatter

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "disallow_int_formatter",
	Doc:      "Disallow fmt int formatter",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

// Check that %d formatting is not used to format pointers in fmt.Sprintf.
// When type of a value changes from `int` to `*int`,
// the resulting format prints memory address instead of value of int.
// See: https://internal-placeholder.com/browse/IDCSTOR-648
// See: https://github.com/golang/go/issues/62595
//
// This check could be more sophisticated: check for argument types and trigger only
// when argument is actually a pointer, track wrapper, etc, with tests on the lint itself,
// but for now just disable `%d` formatter altogether.
// We could take the original go implementation and adapt for our needs,
// see https://cs.opensource.google/go/x/tools/+/master:go/analysis/passes/printf/

func run(pass *analysis.Pass) (interface{}, error) {
	// pass.ResultOf[inspect.Analyzer] will be set if we've added inspect.Analyzer to Requires.
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{ // filter needed nodes: visit only them
		(*ast.CallExpr)(nil),
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		// explore how AST looks via https://astexplorer.net/
		callExpr, ok := node.(*ast.CallExpr)
		if !ok {
			return
		}
		selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		if len(callExpr.Args) < 2 {
			return
		}

		xIdent, ok := selectorExpr.X.(*ast.Ident)
		if !ok {
			return
		}

		// Only process if expression is from "fmt" package
		if xIdent.Name != "fmt" {
			return
		}

		if !detectPrintCode(selectorExpr.Sel.Name) {
			return
		}

		// Check if pattern contains %d
		arg0, ok := callExpr.Args[0].(*ast.BasicLit)
		if !ok || arg0.Kind != token.STRING {
			return
		}
		pattern, err := strconv.Unquote(arg0.Value)
		if err != nil {
			return
		}
		// Todo could be false positive if conatins '%%d'
		if strings.Contains(pattern, "%d") {
			pass.Reportf(arg0.Pos(), "disallowed pattern in format string:\n    %s\n    replace '%%d' pattern with '%%s' and use strconv.Itoa", arg0.Value)
		}

	})

	return nil, nil
}

func detectPrintCode(selName string) bool {
	switch selName {
	case "Appendf", "Sprintf", "Printf", "Errorf":
		return true
	default:
		return false
	}
}

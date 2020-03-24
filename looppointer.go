package looppointer

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:             "looppointer",
	Doc:              "checks for pointers to enclosing loop variables",
	Run:              run,
	RunDespiteErrors: true,
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
	// ResultType reflect.Type
	// FactTypes []Fact
}

func init() {
	//	Analyzer.Flags.StringVar(&v, "name", "default", "description")
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.RangeStmt)(nil),
		(*ast.ForStmt)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		// Find the variables updated by the loop statement.
		var vars []*ast.Ident
		addVar := func(expr ast.Expr) {
			if id, ok := expr.(*ast.Ident); ok {
				vars = append(vars, id)
			}
		}
		var body *ast.BlockStmt
		switch n := n.(type) {
		case *ast.RangeStmt:
			body = n.Body
			addVar(n.Key)
			addVar(n.Value)
		case *ast.ForStmt:
			body = n.Body
			switch post := n.Post.(type) {
			case *ast.AssignStmt:
				// e.g. for p = head; p != nil; p = p.next
				for _, lhs := range post.Lhs {
					addVar(lhs)
				}
			case *ast.IncDecStmt:
				// e.g. for i := 0; i < n; i++
				addVar(post.X)
			}
		}
		if vars == nil {
			return
		}

		// Find the variables declared in the loop.
		var inVars []*ast.Ident
		addInvar := func(expr ast.Expr) {
			if id, ok := expr.(*ast.Ident); ok {
				inVars = append(inVars, id)
			}
		}
		// Inspect internal variables
		ast.Inspect(body, func(n ast.Node) bool {
			switch typed := n.(type) {
			case (*ast.DeclStmt):
				//UNDONE: Var宣言だったらaddInvar
			case (*ast.AssignStmt):
				// Find statements declaring internal variable
				if typed.Tok == token.DEFINE {
					for _, h := range typed.Lhs {
						addInvar(h)
					}
				}
			}
			return true
		})

		// Inspect assigning a pointer to a loop variable to outer var
		ast.Inspect(body, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok || assign.Tok == token.DEFINE {
				return true
			}

			// Find expressions to check
			var checks []ast.Expr
			if len(assign.Rhs) == 1 {
				// If the least one left-hand is NOT a variable in the loop,
				// right-hands may be checked.
			SEARCH_LHS:
				for _, lh := range assign.Lhs {
					leftID, ok := lh.(*ast.Ident)
					if !ok {
						continue
					}
					for _, v := range inVars {
						if v.Obj == leftID.Obj {
							checks = append(checks, assign.Rhs[0])
							break SEARCH_LHS
						}
					}
				}
			} else if len(assign.Rhs) == len(assign.Lhs) {
				// If a left-hand is NOT a variable in the loop,
				// corresponding right-hand may be checked.
				for i, lh := range assign.Lhs {
					leftID, ok := lh.(*ast.Ident)
					if !ok {
						continue
					}
					for _, v := range inVars {
						if v.Obj == leftID.Obj {
							checks = append(checks, assign.Rhs[i])
						}
					}
				}
			}

			// Find pointer to a loop variable
			for _, expr := range checks {
				ast.Inspect(expr, func(n ast.Node) bool {
					switch typed := n.(type) {
					case (*ast.CallExpr):
						// expand append call
						if typed.Fun.(*ast.Ident).Name == "append" {
							return true
						}
						// UNDONE: case struct member assign
					case (*ast.UnaryExpr):
						// find pointer reference
						// 	if assign.Op != token.AND {
						// 		return false
						// 	}
						// 	id, ok := unary.X.(*ast.Ident)
						// 	if !ok {
						// 		return true
						// 	}
						// 	if id.Obj == nil {
						// 		return true
						// 	}
						// if pass.TypesInfo.Types[id].Type == nil {
						// 	// Not referring to a variable (e.g. struct field name)
						// 	return true
						// }
						// for _, v := range vars {
						// 	if v.Obj == id.Obj {
						// 		pass.ReportRangef(id, "loop variable %s captured by a pointer", id.Name)
						// 	}
						// }
					}
					return false
				})
			}
			return true
		})
	})
	return nil, nil
}

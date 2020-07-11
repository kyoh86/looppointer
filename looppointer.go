package looppointer

import (
	"fmt"
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

	search := &Searcher{
		Stats: map[token.Pos]struct{}{},
	}

	nodeFilter := []ast.Node{
		(*ast.RangeStmt)(nil),
		(*ast.ForStmt)(nil),
		(*ast.UnaryExpr)(nil),
	}

	inspect.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		id, digg := search.Check(n, stack)
		if id != nil {
			msg := fmt.Sprintf("taking a pointer for the loop variable %s", id.Name)
			pass.Report(analysis.Diagnostic{Pos: id.Pos(), End: id.End(), Message: msg, Category: "looppointer"})
		}
		return digg
	})

	return nil, nil
}

type Searcher struct {
	// statement variables
	Stats map[token.Pos]struct{}
}

func (s *Searcher) Check(n ast.Node, stack []ast.Node) (*ast.Ident, bool) {
	switch typed := n.(type) {
	case *ast.RangeStmt:
		s.parseRangeStmt(typed)
	case *ast.ForStmt:
		s.parseForStmt(typed)
	case *ast.UnaryExpr:
		return s.checkUnaryExpr(typed, stack)
	}
	return nil, true
}

func (s *Searcher) parseRangeStmt(n *ast.RangeStmt) {
	s.addStat(n.Key)
	s.addStat(n.Value)
}

func (s *Searcher) parseForStmt(n *ast.ForStmt) {
	switch post := n.Post.(type) {
	case *ast.AssignStmt:
		// e.g. for p = head; p != nil; p = p.next
		for _, lhs := range post.Lhs {
			s.addStat(lhs)
		}
	case *ast.IncDecStmt:
		// e.g. for i := 0; i < n; i++
		s.addStat(post.X)
	}
}

func (s *Searcher) addStat(expr ast.Expr) {
	if id, ok := expr.(*ast.Ident); ok {
		s.Stats[id.Pos()] = struct{}{}
	}
}

func (s *Searcher) innermostLoop(stack []ast.Node) ast.Node {
	for i := len(stack) - 1; i >= 0; i-- {
		switch stack[i].(type) {
		case *ast.RangeStmt, *ast.ForStmt:
			return stack[i]
		}
	}
	return nil
}

func (s *Searcher) checkUnaryExpr(n *ast.UnaryExpr, stack []ast.Node) (*ast.Ident, bool) {
	loop := s.innermostLoop(stack)
	if loop == nil {
		return nil, true
	}

	if n.Op != token.AND {
		return nil, true
	}

	// Get identity of the referred item
	id := getIdentity(n.X)
	if id == nil {
		return nil, true
	}

	// If the identity is not the loop statement variable,
	// it will not be reported.
	if _, isStat := s.Stats[id.Obj.Pos()]; !isStat {
		return nil, true
	}

	return id, false
}

// Get variable identity
func getIdentity(expr ast.Expr) *ast.Ident {
	switch typed := expr.(type) {
	case *ast.SelectorExpr:
		// Get parent identity; i.e. `a` of the `a.b`.
		parent, ok := typed.X.(*ast.Ident)
		if !ok {
			return nil
		}
		// NOTE: If that is descendants member like `a.b.c`,
		//       typed.X will be `*ast.SelectorExpr`.
		return parent

	case *ast.Ident:
		// Get simple identity; i.e. `a` of the `a`.
		if typed.Obj == nil {
			return nil
		}
		return typed
	}
	return nil
}

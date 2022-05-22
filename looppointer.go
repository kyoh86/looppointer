package looppointer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/kyoh86/nolint"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:             "looppointer",
	Doc:              "checks for pointers to enclosing loop variables",
	Run:              run,
	RunDespiteErrors: true,
	Requires:         []*analysis.Analyzer{inspect.Analyzer, nolint.Analyzer},
	// ResultType reflect.Type
	// FactTypes []Fact
}

func init() {
	//	Analyzer.Flags.StringVar(&v, "name", "default", "description")
}

type messagesFormats struct {
	dMsgFormat string
	fMsgFormat string
}

const (
	loopPointer   = "looppointer"
	funcReference = "func-reference"
)

var detectionTypeToMessageFormats = map[string]messagesFormats{
	loopPointer: {
		"taking a pointer for the loop variable %s",
		"loop variable %s should be pinned",
	},
	funcReference: {
		"using loop variable in function literal %s",
		"loop variable %s should be pinned",
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	noLinter := pass.ResultOf[nolint.Analyzer].(*nolint.NoLinter)

	search := &Searcher{
		Stats: map[token.Pos]struct{}{},
	}

	nodeFilter := []ast.Node{
		(*ast.RangeStmt)(nil),
		(*ast.ForStmt)(nil),
		(*ast.UnaryExpr)(nil),
		(*ast.Ident)(nil),
	}

	inspect.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		issueType, id, insert, digg := search.Check(n, stack)
		if id != nil && !noLinter.IgnoreNode(id, "looppointer") {
			dMsg := fmt.Sprintf(detectionTypeToMessageFormats[issueType].dMsgFormat, id.Name)
			fMsg := fmt.Sprintf(detectionTypeToMessageFormats[issueType].fMsgFormat, id.Name)
			var suggest []analysis.SuggestedFix
			if insert != token.NoPos {
				suggest = []analysis.SuggestedFix{{
					Message: fMsg,
					TextEdits: []analysis.TextEdit{{
						Pos:     insert,
						End:     insert,
						NewText: []byte(fmt.Sprintf("%[1]s := %[1]s\n", id.Name)),
					}},
				}}
			}
			pass.Report(analysis.Diagnostic{
				Pos:            id.Pos(),
				End:            id.End(),
				Message:        dMsg,
				Category:       "looppointer",
				SuggestedFixes: suggest,
			})
		}
		return digg
	})

	return nil, nil
}

type Searcher struct {
	// statement variables
	Stats map[token.Pos]struct{}
}

func (s *Searcher) Check(n ast.Node, stack []ast.Node) (string, *ast.Ident, token.Pos, bool) {
	switch typed := n.(type) {
	case *ast.RangeStmt:
		s.parseRangeStmt(typed)
	case *ast.ForStmt:
		s.parseForStmt(typed)
	case *ast.UnaryExpr:
		return s.checkUnaryExpr(typed, stack)
	case *ast.Ident:
		return s.checkIdent(typed, stack)
	}
	return "", nil, token.NoPos, true
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

func insertionPosition(block *ast.BlockStmt) token.Pos {
	if len(block.List) > 0 {
		return block.List[0].Pos()
	}
	return token.NoPos
}

func (s *Searcher) innermostLoop(stack []ast.Node) (ast.Node, token.Pos) {
	for i := len(stack) - 1; i >= 0; i-- {
		switch typed := stack[i].(type) {
		case *ast.RangeStmt:
			return typed, insertionPosition(typed.Body)
		case *ast.ForStmt:
			return typed, insertionPosition(typed.Body)
		}
	}
	return nil, token.NoPos
}

func (s *Searcher) checkIdent(n *ast.Ident, stack []ast.Node) (string, *ast.Ident, token.Pos, bool) {
	// Get identity of the referred item
	id := getIdentity(n)
	if id == nil || id.Obj == nil {
		return "", nil, token.NoPos, true
	}

	if _, isStat := s.Stats[id.Obj.Pos()]; isStat {

		// Find any function literals in the stack above
		for i := len(stack) - 1; i >= 0; i-- {
			stackNode := stack[i]

			if fl, ok := stackNode.(*ast.FuncLit); ok {
				// If this is a usage within a function literal, but the variable was declared in this function literal
				// then there's no issue.
				if n.Obj.Pos() > fl.Pos() && n.Obj.Pos() < fl.End() {
					return "", nil, n.Pos(), true
				}

				return funcReference, n, n.Pos(), false
			}
		}

	}

	return "", nil, n.Pos(), true
}

func (s *Searcher) checkUnaryExpr(n *ast.UnaryExpr, stack []ast.Node) (string, *ast.Ident, token.Pos, bool) {
	loop, insert := s.innermostLoop(stack)
	if loop == nil {
		return "", nil, token.NoPos, true
	}

	if n.Op != token.AND {
		return "", nil, token.NoPos, true
	}

	// Get identity of the referred item
	id := getIdentity(n.X)
	if id == nil || id.Obj == nil {
		return "", nil, token.NoPos, true
	}

	// If the identity is not the loop statement variable,
	// it will not be reported.
	if _, isStat := s.Stats[id.Obj.Pos()]; !isStat {
		return "", nil, token.NoPos, true
	}

	return loopPointer, id, insert, false
}

// Get variable identity
func getIdentity(expr ast.Expr) *ast.Ident {
	switch typed := expr.(type) {
	case *ast.SelectorExpr:
		// Get parent identity
		// i.e.
		// `a` of the `a.b`.
		// `a.b` of the `a.b.c`.
		return getIdentity(typed.X)

	case *ast.Ident:
		// Get simple identity; i.e. `a` of the `a`.
		if typed.Obj == nil {
			return nil
		}
		return typed
	}
	return nil
}

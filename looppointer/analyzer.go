package looppointer

import (
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name:             "looppointer",
	Doc:              "checks for pointers to enclosing loop variables",
	Run:              run,
	RunDespiteErrors: true,
	// ResultType reflect.Type
	// FactTypes []Fact
	// Requires []*Analyzer
}

func init() {
	//	Analyzer.Flags.StringVar(&v, "name", "default", "description")
}

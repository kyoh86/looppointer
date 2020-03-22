package main

import (
	"github.com/kyoh86/looppointer/looppointer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(looppointer.Analyzer)
}

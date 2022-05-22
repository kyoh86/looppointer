package looppointer_test

import (
	"testing"

	"github.com/kyoh86/looppointer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, looppointer.Analyzer, "simple")
	analysistest.Run(t, testdata, looppointer.Analyzer, "fixed")
	analysistest.Run(t, testdata, looppointer.Analyzer, "issue7")
	analysistest.Run(t, testdata, looppointer.Analyzer, "nolint")
	analysistest.Run(t, testdata, looppointer.Analyzer, "nested")
	analysistest.Run(t, testdata, looppointer.Analyzer, "embedded-func-pointer")
}

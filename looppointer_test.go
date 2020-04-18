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
}

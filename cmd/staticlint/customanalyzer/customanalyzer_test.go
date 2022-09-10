// Package customanalyzer provides custom code analysis.
package customanalyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestMyAnalyzer(t *testing.T) {
	// function analysistest.Run applies OsExitInMainAnalyzer to packages from testdata and checks expected result
	analysistest.Run(t, analysistest.TestData(), OsExitInMainAnalyzer, "./...")
}

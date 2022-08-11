// Package customanalyzer provides custom code analysis.
package customanalyzer

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func TestMyAnalyzer(t *testing.T) {
	// функция analysistest.Run применяет тестируемый анализатор OsExitInMainAnalyzer
	// к пакетам из папки testdata и проверяет ожидания
	analysistest.Run(t, analysistest.TestData(), OsExitInMainAnalyzer, "./...")
}
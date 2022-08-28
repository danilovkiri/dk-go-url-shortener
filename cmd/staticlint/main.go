package main

import (
	"github.com/danilovkiri/dk_go_url_shortener/cmd/staticlint/customanalyzer"
	"github.com/gostaticanalysis/sqlrows/passes/sqlrows"
	"github.com/reillywatson/lintservemux"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"honnef.co/go/tools/staticcheck"
	"strings"
)

func main() {
	// passesChecks contains selected "golang.org/x/tools/go/analysis/passes" analyzers
	passesChecks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		assign.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stringintconv.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
	}
	// staticChecks contains selected "honnef.co/go/tools/staticcheck" analyzers
	var staticChecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Name, "SA") || strings.HasPrefix(v.Name, "ST") {
			staticChecks = append(staticChecks, v)
		}
	}
	// publicChecks contains selected publicly available analyzers
	publicChecks := []*analysis.Analyzer{
		sqlrows.Analyzer,
		lintservemux.Analyzer,
	}
	// customChecks contains custom analyzers
	customChecks := []*analysis.Analyzer{
		customanalyzer.OsExitInMainAnalyzer,
	}
	// running analyzers
	allChecks := append(passesChecks, staticChecks...)
	allChecks = append(allChecks, publicChecks...)
	allChecks = append(allChecks, customChecks...)
	multichecker.Main(
		allChecks...,
	)

}

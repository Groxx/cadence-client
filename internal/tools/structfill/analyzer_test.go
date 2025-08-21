package structfill

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "example.org/structfill/basic")
}

func TestEnforcerMode(t *testing.T) {
	testdata := analysistest.TestData()
	// Test with enforcer mode enabled for the "enforced" package
	analyzer := *Analyzer
	analyzer.Flags.Set("enforce", "example.org/structfill/enforced")
	analysistest.Run(t, testdata, &analyzer, "example.org/structfill/enforced")
}

func TestIgnoreTests(t *testing.T) {
	testdata := analysistest.TestData()
	// Test with ignore-tests flag enabled - should not report any diagnostics on test files
	analyzer := *Analyzer
	analyzer.Flags.Set("ignore-tests", "true")
	analyzer.Flags.Set("ignore-generated", "true")
	analysistest.Run(t, testdata, &analyzer, "example.org/structfill/ignore_tests")
}

func TestEnforcerModeWithIgnoreGenerated(t *testing.T) {
	testdata := analysistest.TestData()
	// Test with both enforcer mode AND ignore-generated enabled
	// Should enforce struct filling in enforced package but skip generated files
	analyzer := *Analyzer
	analyzer.Flags.Set("enforce", "example.org/structfill/enforced_generated")
	analyzer.Flags.Set("ignore-generated", "true")
	analysistest.Run(t, testdata, &analyzer, "example.org/structfill/enforced_generated")
}

func TestStructLevelCanSkip(t *testing.T) {
	testdata := analysistest.TestData()
	// Test that struct-level // lint:can-skip exempts structs from enforcer mode
	// while struct-level // lint:must-fill still enforces rules
	analyzer := *Analyzer
	analyzer.Flags.Set("enforce", "example.org/structfill/enforced")
	analysistest.Run(t, testdata, &analyzer, "example.org/structfill/enforced")
}

func TestSkipSpecificTypes(t *testing.T) {
	testdata := analysistest.TestData()
	// Test that -skip flag can exempt specific types from enforcement
	analyzer := *Analyzer
	analyzer.Flags.Set("skip", "example.org/structfill/skip_types.SkippableType")
	analysistest.Run(t, testdata, &analyzer, "example.org/structfill/skip_types")
}

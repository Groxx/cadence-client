package xaccess

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()

	LimitAccess("example.org/decl")
	LimitAccess("example.org/decl.Func", "example.org/allowed")
	LimitAccess("example.org/decl.Thing", "example.org/allowed")
	func() {
		defer func() { recover() }()
		LimitAccess("example.org/decl.Func")
		t.Error("should have panicked when trying to restrict a package that already allows some things")
	}()
	t.Cleanup(func() {
		LimitAccess("")
	})

	analysistest.Run(t, testdata, Analyzer, "example.org/...")
}

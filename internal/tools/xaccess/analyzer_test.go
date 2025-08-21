package xaccess

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	ControlledAccess = Config{
		// does not support ... wildcards so no parsing needs to happen, just equality.
		// if you need this, you can just build your list from `go list`, and it'll always be correct.
		"example.org/decl": {},
		// specific-thing definitions are additive, and can allow stuff that is blocked at a package level.
		// decl.Public is still blocked.
		"example.org/decl.Func": {
			Allowed: map[string]bool{
				"example.org/allowed": true,
			},
		},
	}
	t.Cleanup(func() {
		ControlledAccess = nil
	})
	analysistest.Run(t, testdata, Analyzer, "example.org/...")
}

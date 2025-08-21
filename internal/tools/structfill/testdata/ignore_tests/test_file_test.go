package structfill

import "testing"

// This test file should be ignored when -ignore-tests=true is set
// Even though this struct has no lint:must-fill comment and there are
// incomplete struct literals, they should be ignored in test files

type TestStruct struct {
	Field1 string
	Field2 int
}

// These incomplete struct literals should be ignored when -ignore-tests=true
var testCases = []TestStruct{
	{Field1: "test1"}, // Missing Field2, but should be ignored in tests
	{Field2: 42},      // Missing Field1, but should be ignored in tests
}

func TestSomething(t *testing.T) {
	// This incomplete struct should be ignored when -ignore-tests=true
	ts := TestStruct{Field1: "incomplete"} // Missing Field2
	_ = ts
}

package structfill

// lint:must-fill
type SkippableType struct {
	Field1 string
	Field2 int
}

// When -skip flag is used for this type, this should NOT trigger an error
// (The test will run with -skip="example.org/structfill.SkippableType")
var skippedExample = SkippableType{
	Field1: "test",
	// Field2 missing, but should be ignored due to -skip flag
}

// This should still be valid when all fields are filled
var completeExample = SkippableType{
	Field1: "test",
	Field2: 42,
}

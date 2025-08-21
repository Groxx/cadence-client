package enforced

// This struct is in the enforced package but has // lint:can-skip at struct level
// so it should be exempt from the enforcement rules
// lint:can-skip this struct is exempt from enforcer mode
type ExemptedStruct struct {
	Field1 string
	Field2 int
}

// These incomplete struct literals should be allowed since the struct is exempted
var exemptedExamples = []ExemptedStruct{
	{Field1: "test1"}, // Missing Field2, but allowed since struct is exempted
	{Field2: 42},      // Missing Field1, but allowed since struct is exempted
}

// This struct also in enforced package but with explicit must-fill should still be enforced
// lint:must-fill
type ExplicitStruct struct { // want ExplicitStruct:""
	Name string
	Age  int
}

// This should still trigger errors since it has explicit must-fill
var explicitExample = ExplicitStruct{ // want `missing "Age"`
	Name: "test",
}

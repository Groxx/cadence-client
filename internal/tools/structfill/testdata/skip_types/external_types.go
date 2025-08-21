package structfill

// lint:must-fill
type ExternalType struct { // want ExternalType:""
	Field1 string
	Field2 int
}

// This struct literal should normally trigger an error for missing Field2
var badExternal = ExternalType{ // want `missing "Field2"`
	Field1: "test",
}

// This should be valid when all fields are filled
var goodExternal = ExternalType{
	Field1: "test",
	Field2: 42,
}
package structfill

// some comments
// lint:must-fill
type Thing struct { // want Thing:""
	Field  string
	Field2 string
	// lint:can-skip
	OptionalField string
}

// this does not have the magic comment,
// so missing fields are allowed
type Thing2 struct {
	Field  string
	Field2 string
}

// lint:must-fill
type Embedded struct { // want Embedded:""
	Thing
	Other string
}

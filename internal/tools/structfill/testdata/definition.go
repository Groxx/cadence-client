package structfill

// some comments
// lint:must-fill
type Thing struct { // want Thing:""
	Field  string
	Field2 string
	// lint:can-skip this field is optional for backwards compatibility
	OptionalField string
}

// this does not have the magic comment,
// so missing fields are allowed
type Thing2 struct {
	Field  string
	Field2 string
}

// lint:must-fill with descriptive text allowed
type Embedded struct { // want Embedded:""
	Thing
	Other string
}

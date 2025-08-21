package structfill

var _ = Thing{ // want `missing "field2"`
	Field: "foo",
	// OptionalField can be omitted thanks to lint:can-skip
}

// This should be valid - OptionalField can be skipped
var _ = Thing{
	Field:  "foo",
	Field2: "bar",
}

// This should also be valid - OptionalField is explicitly filled
var _ = Thing{
	Field:         "foo",
	Field2:        "bar",
	OptionalField: "baz",
}

var _ = Thing2{
	Field: "foo",
	// Field2: "", not filling it is allowed
}

// getting a pointer to a struct must still be filled
var _ = &Thing{ // want `missing "field2"`
	Field: "foo",
}

var _ = struct {
	Field  string
	Field2 string
}{
	Field: "foo",
}
var (
	_ = Embedded{ // want `missing "thing"`
		Other: "bar",
	}

	_ = Embedded{
		Thing: Thing{}, // want `missing "field"` `missing "field2"`
		Other: "bar",
	}
)

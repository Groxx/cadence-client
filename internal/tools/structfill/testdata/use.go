package structfill

var _ = Thing{ // want `missing "field2"`
	Field: "foo",
}

var _ = Thing2{
	Field: "foo",
	// Field2: "", not filling it is allowed
}

var _ = struct {
	Field  string
	Field2 string
}{
	Field: "foo",
}

var _ = Embedded{ // want `missing "thing"`
	Other: "bar",
}

var _ = Embedded{
	Thing: Thing{}, // want `missing "field"` `missing "field2"`
	Other: "bar",
}

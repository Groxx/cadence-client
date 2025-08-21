package enforced

// When enforcer mode is enabled for this package, these should trigger errors
var _ = APIRequest{ // want `missing "URL"`
	Method: "GET",
	// LegacyField can be skipped due to // lint:can-skip comment
}

var _ = APIResponse{ // want `missing "Body"`
	Status: 200,
}

// These should be valid when all fields are filled
var _ = APIRequest{
	Method: "POST",
	URL:    "/api/test",
	// LegacyField omitted - OK due to lint:can-skip
}

var _ = APIResponse{
	Status: 201,
	Body:   "Created",
}

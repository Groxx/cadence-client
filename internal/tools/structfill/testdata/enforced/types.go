package enforced

// This struct has no // lint:must-fill comment, but should be enforced
// when the package is in enforcer mode
type APIRequest struct { // want APIRequest:""
	Method string
	URL    string
	// lint:can-skip for backwards compatibility
	LegacyField string
}

// This struct also has no comment but should be enforced
type APIResponse struct { // want APIResponse:""
	Status int
	Body   string
}

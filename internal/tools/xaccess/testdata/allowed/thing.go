package allowed

import thing "example.org/decl"

var _ = thing.Func // allowed, so it should be silent

// not in the allowed list
var _ = thing.Public // want `using controlled object`

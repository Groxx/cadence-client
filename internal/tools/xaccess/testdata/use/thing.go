package use

import thing "example.org/decl"

var x = thing.Func // want `using controlled object`

func init() {
	_ = x        // should be silent
	thing.Func() // want `using controlled object`
}

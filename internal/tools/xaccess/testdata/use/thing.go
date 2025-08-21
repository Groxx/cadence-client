package use

import thing "example.org/decl"

var x = thing.Func // want `using controlled object`

func init() {
	_ = x        // should be silent, only the reference itself is reported
	thing.Func() // want `using controlled object`
}

package allowed

import thing "example.org/decl"

var _ = thing.Func // allowed, so it should be silent

// not in the allowed list
var _ = thing.Public // want `using controlled object`

// Thing is allowed, but the embedded field is not.
// This is currently allowed because... well because it's easy.
//
// I think it's defensible: the embedded field is reasonably part of "Thing" itself.
// And it's not really possible to block anyway - embedding it in something requires
// access to it, at which point you can just return it via an `any`.
var _ = thing.Thing{}.Public          // allowed, so it should be silent
var _ = thing.Thing{}.Embedded.Public // much more dubious, but also allowed, it's "just" a "field"
// using the embedded thing by type is still blocked, though.
var _ = thing.Embedded{}.Public // want `using controlled object`

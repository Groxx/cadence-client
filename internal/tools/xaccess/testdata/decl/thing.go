package thing

func Func() {} // want Func:""

func Public() { // want Public:""
	type Internal struct{} // not accessible, no facts
	_ = Internal{}
	Func() // in-package use, allowed
}

func private() {} // not accessible, no facts

type Thing struct { // want Thing:""
	Field string
	Embedded
}
type Embedded struct { // want Embedded:""
	Public string
}

func (t Thing) Method() {} // want Method:""

type Thingiface interface { // want Thingiface:""
	Method() string
}

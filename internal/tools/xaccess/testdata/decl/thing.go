package thing

func Func() {} // want Func:""

func Public() { // want Public:""
	type Internal struct{} // not accessible, no facts
	_ = Internal{}
}

func private() {} // not accessible, no facts

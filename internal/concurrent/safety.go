// Copyright (c) 2017-2021 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package concurrent

import "math"

// NoCopy looks like a mutex, which causes `go vet` to check for copies.
//
// to use, just add into a container that must not be copied, e.g. as an un-named field.
// it does not take any space, it's just a compile-time construct.
type NoCopy struct{}

func (*NoCopy) Lock()   {}
func (*NoCopy) Unlock() {}

// NoCmp is a zero-sized non-comparable type, which will have compile-time failures on comparison checks.
//
// to use, just add it into a container that must not be compared, e.g. as an un-named field.
// it does not take any space, it's just a compile-time construct.
type NoCmp [0]func()

// compile-time guarantee of 64-bit architecture, which ensures int casts are correct in AtomicInt.
// to see a failure, try `GOOS=linux GOARCH=386 make build`.
//
// this could also be guaranteed by making this file buildable only on 64-bit arches,
// but this tactic will work on all future 64-bit architectures too.
var _ [math.MaxInt64]struct{} = (func() (res [math.MaxInt]struct{}) { return })()

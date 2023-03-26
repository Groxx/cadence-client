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

import (
	"sync"
)

// Locker wraps a variable with a mutex and accessors.
type Locker[Value any, Mutable *Value] struct {
	value Value
	mut   sync.Mutex // contains a NoCopy

	// mutex does not contain a no-compare marker.
	// which is valid, but mutex-containing things are often also unsafe to compare.
	_ NoCmp
}

// NewLocker initializes and returns a Locker, which is a wrapper around a value-type
// and a mutex.
//
// Value-returns are safe to access outside the lock, as long as the Value type is actually
// a value-type, and is not mutable.
//
// Mutable-returns can be updated with `*val = newval`, but all reads and writes MUST occur
// while the lock is held.  For safety's sake, try to do this via Locker.Change.
//
// TODO: an Analysis-based linter could assert that Value is only used for value-types.
func NewLocker[Value any, Ptr *Value](initial Value) *Locker[Value, Ptr] {
	return &Locker[Value, Ptr]{value: initial}
}

// Lock is a low-level mutex-like accessor, and it returns a mutable reference.
// Whenever possible, use Change instead.
//
// It is the caller's responsibility to ensure the mut-return is not accessed in any way
// after the unlock func is called.  You are encouraged to set your variable to nil before
// unlocking to prevent accidental leaks.
//
// E.g. do this:
//
//	// good
//	val, unlock := l.Lock()
//	if val == asdf {
//	  // do stuff
//	  *val = newValue   // replace the value
//	  val.Field = other // or otherwise mutate it
//	}
//	val = nil // forget the reference for safety
//	unlock()  // release the lock
//	// val is nil and cannot be misused here
//
// But do not do this:
//
//	// bad
//	val, unlock := l.Lock()
//	val.Field = newValue         // update it (safe)
//	unlock()                     // release the lock
//	*val = asdf                  // !! UNSAFE DATA RACE !!
//	if val.Field == asdf { ... } // !! UNSAFE DATA RACE !!
//	                             // The value or Field may have been mutated by another goroutine.
func (l *Locker[Value, Mutable]) Lock() (mut Mutable, unlock func()) {
	l.mut.Lock()
	return &l.value, l.mut.Unlock
}

// Change runs the callback function while holding the lock, and ensures the lock
// is released before returning.
//
// As long as you do not leak a reference to the Mutable value, this should be safe from misuse.
func (l *Locker[Value, Mutable]) Change(fn func(value Mutable)) (newValue Value) {
	l.mut.Lock()
	defer l.mut.Unlock()

	fn(&l.value)
	return l.value
}

// Get retrieves the current Value of this Locker.
//
// Since Value is a value-type, the returned value is safe to read at any time.
func (l *Locker[Value, Mutable]) Get() Value {
	l.mut.Lock()
	defer l.mut.Unlock()
	return l.value
}

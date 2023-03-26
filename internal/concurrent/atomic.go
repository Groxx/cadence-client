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
	"sync/atomic"
)

// AtomicInt holds and accesses an int (technically an int64) atomically.
// it must be used as a pointer.
type AtomicInt struct {
	value int64

	_ NoCopy // copies by the runtime do not use atomics, so they violate memory safety
	_ NoCmp  // same as NoCopy
}

func NewAtomicInt(initial int) *AtomicInt {
	return &AtomicInt{value: int64(initial)}
}

func (a *AtomicInt) Load() int {
	return int(atomic.LoadInt64(&a.value))
}
func (a *AtomicInt) Inc() int {
	return a.Add(1)
}
func (a *AtomicInt) Add(delta int) int {
	return int(atomic.AddInt64(&a.value, int64(delta)))
}
func (a *AtomicInt) Store(value int) {
	atomic.StoreInt64(&a.value, int64(value))
}
func (a *AtomicInt) CAS(old, new int) (swapped bool) {
	return atomic.CompareAndSwapInt64(&a.value, int64(old), int64(new))
}

// AtomicBool holds and accesses a bool (technically an int64) atomically.
// it must be used as a pointer.
type AtomicBool struct {
	value int64 // 0 is false, 1 is true

	_ NoCopy // copies by the runtime do not use atomics, so they violate memory safety
	_ NoCmp  // same as NoCopy
}

func NewAtomicBool(initial bool) *AtomicBool {
	a := AtomicBool{value: 0}
	if initial {
		a.value = 1
	}
	return &a
}
func (a *AtomicBool) Load() bool {
	return atomic.LoadInt64(&a.value) == 1
}
func (a *AtomicBool) Store(value bool) {
	var bnew int64
	if value {
		bnew = 1
	}
	atomic.StoreInt64(&a.value, bnew)
}
func (a *AtomicBool) Swap(value bool) (old bool) {
	var bnew int64
	if value {
		bnew = 1
	}
	prev := atomic.SwapInt64(&a.value, bnew)
	return prev == 1
}

func (a *AtomicBool) CAS(old, new bool) (swapped bool) {
	var bold, bnew int64
	if old {
		bold = 1
	}
	if new {
		bnew = 1
	}
	return atomic.CompareAndSwapInt64(&a.value, bold, bnew)
}

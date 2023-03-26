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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicInt(t *testing.T) {
	// this should be sufficient for all atomics
	var startWg sync.WaitGroup
	var doneWg sync.WaitGroup
	race := 100
	startWg.Add(race + 1)
	doneWg.Add(race)
	val := NewAtomicInt(0)
	for i := 0; i < race; i++ {
		go (func() {
			startWg.Done() // mark this goroutine as started
			startWg.Wait() // wait for all goroutines to start

			val.Add(1)
			doneWg.Done()
		})()
	}

	startWg.Done() // all goroutines created
	doneWg.Wait()  // wait for all goroutines to finish
	assert.Equal(t, race, val.Load())
}

func TestAtomicBool(t *testing.T) {
	// this should be sufficient for all atomics
	var startWg sync.WaitGroup
	var doneWg sync.WaitGroup
	race := 100
	startWg.Add(race + 1)
	doneWg.Add(race)
	val := NewAtomicBool(false)
	for i := 0; i < race; i++ {
		go (func() {
			startWg.Done() // mark this goroutine as started
			startWg.Wait() // wait for all goroutines to start

			val.Store(i%2 == 1)
			doneWg.Done()
		})()
	}

	startWg.Done() // all goroutines created
	doneWg.Wait()  // wait for all goroutines to finish
	_ = val.Load() // in case it's necessary to defeat optimizations
}
